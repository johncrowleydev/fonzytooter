package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/auth"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
	"github.com/johncrowleydev/fonzytooter/server/internal/review"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
)

type reviewTestClock struct{ now time.Time }

func (clock reviewTestClock) Now() time.Time { return clock.now }

func TestReviewAPIListsVirtualCardsAndCreatesRating(t *testing.T) {
	app, db := testReviewAPI(t)

	response := serve(t, app.Handler, http.MethodGet, "/api/courses/course/review-cards?due=true")
	if response.Code != http.StatusOK {
		t.Fatalf("list review cards status = %d: %s", response.Code, response.Body.String())
	}
	var cards []ReviewCardResource
	decodeJSON(t, response, &cards)
	if len(cards) != 1 || !cards[0].Virtual || cards[0].State != "new" || len(cards[0].Previews) != 4 {
		t.Fatalf("unexpected virtual card: %#v", cards)
	}
	assertHTTPTableCount(t, db, "review_cards", 0)

	response = serveJSON(t, app.Handler, http.MethodPost, "/api/courses/course/modules/module/review-cards/first/reviews", `{"rating":"good"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("create review status = %d: %s", response.Code, response.Body.String())
	}
	if response.Header().Get("Location") == "" {
		t.Fatal("create review omitted Location header")
	}
	var result ReviewCardResource
	decodeJSON(t, response, &result)
	if result.Virtual || result.State == "new" || result.LastReviewedAt == nil {
		t.Fatalf("unexpected resulting review card: %#v", result)
	}
	assertHTTPTableCount(t, db, "review_cards", 1)
	assertHTTPTableCount(t, db, "review_logs", 1)
	assertHTTPTableCount(t, db, "activities", 1)

	response = serve(t, app.Handler, http.MethodGet, "/api/courses/course/review-cards?due=true")
	decodeJSON(t, response, &cards)
	if len(cards) != 0 {
		t.Fatalf("expected no due cards after Good rating, got %#v", cards)
	}
}

func TestReviewAPIRejectsMalformedAndMissingResources(t *testing.T) {
	app, db := testReviewAPI(t)
	tests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{http.MethodPost, "/api/courses/course/modules/module/review-cards/first/reviews", `{"rating":"perfect"}`, http.StatusUnprocessableEntity},
		{http.MethodPost, "/api/courses/course/modules/module/review-cards/missing/reviews", `{"rating":"good"}`, http.StatusNotFound},
		{http.MethodPost, "/api/courses/course/modules/missing/review-cards/first/reviews", `{"rating":"good"}`, http.StatusNotFound},
		{http.MethodPost, "/api/courses/missing/modules/module/review-cards/first/reviews", `{"rating":"good"}`, http.StatusNotFound},
		{http.MethodGet, "/api/courses/missing/review-cards?due=true", "", http.StatusNotFound},
	}
	for _, test := range tests {
		var responseStatus int
		if test.body == "" {
			responseStatus = serve(t, app.Handler, test.method, test.path).Code
		} else {
			responseStatus = serveJSON(t, app.Handler, test.method, test.path, test.body).Code
		}
		if responseStatus != test.status {
			t.Fatalf("%s %s status = %d, want %d", test.method, test.path, responseStatus, test.status)
		}
	}
	assertHTTPTableCount(t, db, "review_cards", 0)
	assertHTTPTableCount(t, db, "review_logs", 0)
}

func TestReviewOpenAPIContract(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), curriculum.NewEmptyCatalog(), nil, nil)
	collection := app.Spec.Paths["/api/courses/{courseId}/review-cards"]
	if collection == nil || collection.Get == nil || collection.Get.OperationID != "listReviewCards" {
		t.Fatalf("missing review-card collection: %#v", collection)
	}
	if collection.Post != nil {
		t.Fatal("review-card collection exposes unintended POST")
	}
	reviews := app.Spec.Paths["/api/courses/{courseId}/modules/{moduleId}/review-cards/{reviewItemId}/reviews"]
	if reviews == nil || reviews.Post == nil || reviews.Post.OperationID != "createReviewCardReview" {
		t.Fatalf("missing subordinate review creation: %#v", reviews)
	}
	if reviews.Get != nil {
		t.Fatal("review creation path exposes unintended GET")
	}
	if app.Spec.Components.Schemas.Map()["ReviewCardList"] == nil {
		t.Fatal("missing generated review-card list component schema")
	}
}

func testReviewAPI(t *testing.T) (*API, *sql.DB) {
	t.Helper()
	files := fstest.MapFS{
		"sources.yaml":                                     {Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                       {Data: []byte("id: course\ntitle: Course\ndescription: Test course.\norder: 0\n")},
		"courses/course/modules/module/module.yaml":        {Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective\n    title: Objective\n    description: Test objective.\n    prerequisites: []\nvideos: []\nlessons:\n  - lesson\n")},
		"courses/course/modules/module/lesson.mdx":         {Data: []byte("---\nid: lesson\ntitle: Lesson\nobjectiveIds:\n  - objective\nsourceIds: []\n---\n# Lesson\n")},
		"courses/course/modules/module/reviews/first.yaml": {Data: []byte("id: first\norder: 0\nobjectiveIds:\n  - objective\nsourceLessonId: lesson\nprompt: What is the neutral answer?\nanswer: The neutral answer.\nhint: Recall the neutral fixture.\n")},
	}
	catalog, err := curriculum.Load(files)
	if err != nil {
		t.Fatalf("load review curriculum: %v", err)
	}
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "learner.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	clock := reviewTestClock{now: time.Date(2026, time.August, 20, 14, 30, 0, 0, time.UTC)}
	if _, err := db.Exec(`
		INSERT INTO lesson_progress (user_id, course_id, module_id, lesson_id, completed, completed_at, updated_at)
		VALUES (?, 'course', 'module', 'lesson', 1, ?, ?)
	`, auth.BootstrapUserID, clock.now.Format(time.RFC3339Nano), clock.now.Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("complete source lesson: %v", err)
	}
	service := review.NewService(db, catalog, clock)
	return newAuthenticatedTestAPI(t, tutor.NewService(tutor.NewUnavailableProvider()), catalog, nil, service), db
}

func assertHTTPTableCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("%s count = %d, want %d", table, got, want)
	}
}
