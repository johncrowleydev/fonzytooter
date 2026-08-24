package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/johncrowleydev/fonzytooter/server/internal/auth"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
	"github.com/johncrowleydev/fonzytooter/server/internal/learner"
	"github.com/johncrowleydev/fonzytooter/server/internal/review"
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
	"golang.org/x/crypto/bcrypt"
)

func TestEveryOperationHasExplicitAccessPolicyAndOpenAPISecurity(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil)
	want := map[string]string{
		"getCurrentAuthenticationSession":    accessPublic,
		"createAuthenticationSession":        accessPublic,
		"deleteCurrentAuthenticationSession": accessPublic,
		"getHealth":                          accessPublic,
		"listCourses":                        accessPublic,
		"getCourse":                          accessPublic,
		"getCourseModule":                    accessPublic,
		"getCourseLesson":                    accessPublic,
		"getCourseModuleExercise":            accessPublic,
		"getCourseModuleReviewItem":          accessPublic,
		"getCourseModuleWorksheet":           accessPublic,
		"getCourseModuleWorksheetDocument":   accessPublic,
		"getCourseModuleWorkbook":            accessPublic,
		"getLessonProgress":                  accessAuthenticated,
		"putLessonProgress":                  accessAuthenticated,
		"getVideoProgress":                   accessAuthenticated,
		"putVideoProgress":                   accessAuthenticated,
		"getCourseProgress":                  accessAuthenticated,
		"listActivities":                     accessAuthenticated,
		"getExerciseCheckDefinition":         accessAuthenticated,
		"getExerciseWorkspace":               accessAuthenticated,
		"putExerciseWorkspace":               accessAuthenticated,
		"createExerciseAttempt":              accessAuthenticated,
		"listReviewCards":                    accessAuthenticated,
		"createReviewCardReview":             accessAuthenticated,
		"createTutorTurn":                    accessAuthenticated,
		"getTutorAccess":                     accessAuthenticated,
	}
	operations := allOperations(app.Spec)
	if len(operations) != len(want) {
		t.Fatalf("classified %d operations, OpenAPI contains %d", len(want), len(operations))
	}
	for operationID, operation := range operations {
		policy, ok := operation.Metadata[accessPolicyMetadataKey].(string)
		if !ok {
			t.Fatalf("operation %s has no access policy", operationID)
		}
		if policy != want[operationID] {
			t.Fatalf("operation %s policy = %q, want %q", operationID, policy, want[operationID])
		}
		if policy == accessPublic {
			if len(operation.Security) != 0 {
				t.Fatalf("public operation %s declares security: %#v", operationID, operation.Security)
			}
			continue
		}
		if len(operation.Security) != 1 || operation.Security[0][sessionSecurityScheme] == nil {
			t.Fatalf("authenticated operation %s has wrong security declaration: %#v", operationID, operation.Security)
		}
		if operation.Responses["401"] == nil {
			t.Fatalf("authenticated operation %s does not document 401", operationID)
		}
	}
	scheme := app.Spec.Components.SecuritySchemes[sessionSecurityScheme]
	if scheme == nil || scheme.Type != "apiKey" || scheme.In != "cookie" || scheme.Name != auth.DefaultCookieName {
		t.Fatalf("unexpected session security scheme: %#v", scheme)
	}
}

func TestUnclassifiedOperationFailsClosed(t *testing.T) {
	app := NewAPI(tutor.NewService(tutor.NewUnavailableProvider()), testCatalog(t), nil, nil)
	app.Spec.Paths["/api/health"].Get.Metadata = nil
	response := authorizationRequest(t, app.Handler, http.MethodGet, "/api/health", "", nil)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("unclassified operation status = %d, want 500: %s", response.Code, response.Body.String())
	}
}

func TestAnonymousAndAuthenticatedAPIBoundary(t *testing.T) {
	app, cookie, stateDB := authorizationTestAPI(t)
	publicPaths := []string{
		"/api/health",
		"/api/courses",
		"/api/courses/ai-ml",
		"/api/courses/ai-ml/modules/python",
		"/api/courses/ai-ml/modules/python/lessons/lesson.stable",
		"/api/courses/ai-ml/modules/python/exercises/python.example",
		"/api/courses/ai-ml/modules/python/review-items/functions.definition",
		"/api/courses/ai-ml/modules/python/worksheets/worksheet",
	}
	for _, path := range publicPaths {
		if response := authorizationRequest(t, app.Handler, http.MethodGet, path, "", nil); response.Code != http.StatusOK {
			t.Fatalf("anonymous public GET %s = %d: %s", path, response.Code, response.Body.String())
		}
	}
	for _, table := range []string{"lesson_progress", "video_progress", "activities", "exercise_workspaces", "exercise_attempts", "review_cards", "review_logs", "tutor_conversations"} {
		var count int
		if err := stateDB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("anonymous curriculum reads changed %s: count=%d err=%v", table, count, err)
		}
	}

	privateRequests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/courses/ai-ml/modules/python/lessons/lesson.stable/progress", ""},
		{http.MethodPut, "/api/courses/ai-ml/modules/python/lessons/lesson.stable/progress", `{"completed":true}`},
		{http.MethodGet, "/api/courses/ai-ml/modules/python/videos/python-video/progress", ""},
		{http.MethodPut, "/api/courses/ai-ml/modules/python/videos/python-video/progress", `{"completed":true}`},
		{http.MethodGet, "/api/courses/ai-ml/progress", ""},
		{http.MethodGet, "/api/activities?courseId=ai-ml", ""},
		{http.MethodGet, "/api/courses/ai-ml/modules/python/exercises/python.example/check-definition", ""},
		{http.MethodGet, "/api/courses/ai-ml/modules/python/exercises/python.example/workspace", ""},
		{http.MethodPut, "/api/courses/ai-ml/modules/python/exercises/python.example/workspace", `{"code":"pass"}`},
		{http.MethodPost, "/api/courses/ai-ml/modules/python/exercises/python.example/attempts", `{}`},
		{http.MethodGet, "/api/courses/ai-ml/review-cards", ""},
		{http.MethodPost, "/api/courses/ai-ml/modules/python/review-cards/functions.definition/reviews", `{}`},
		{http.MethodPost, "/api/tutor/turns", `{"message":"hello"}`},
		{http.MethodGet, "/api/tutor-access", ""},
	}
	for _, request := range privateRequests {
		response := authorizationRequest(t, app.Handler, request.method, request.path, request.body, nil)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("anonymous %s %s = %d, want 401: %s", request.method, request.path, response.Code, response.Body.String())
		}
	}

	authenticatedRequests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{http.MethodGet, "/api/courses/ai-ml/modules/python/lessons/lesson.stable/progress", "", http.StatusOK},
		{http.MethodPut, "/api/courses/ai-ml/modules/python/lessons/lesson.stable/progress", `{"completed":true}`, http.StatusOK},
		{http.MethodGet, "/api/courses/ai-ml/modules/python/videos/python-video/progress", "", http.StatusOK},
		{http.MethodPut, "/api/courses/ai-ml/modules/python/videos/python-video/progress", `{"completed":true}`, http.StatusOK},
		{http.MethodGet, "/api/courses/ai-ml/progress", "", http.StatusOK},
		{http.MethodGet, "/api/activities?courseId=ai-ml", "", http.StatusOK},
		{http.MethodGet, "/api/courses/ai-ml/modules/python/exercises/python.example/check-definition", "", http.StatusOK},
		{http.MethodGet, "/api/courses/ai-ml/modules/python/exercises/python.example/workspace", "", http.StatusOK},
		{http.MethodPut, "/api/courses/ai-ml/modules/python/exercises/python.example/workspace", `{"code":"pass"}`, http.StatusOK},
		{http.MethodPost, "/api/courses/ai-ml/modules/python/exercises/python.example/attempts", `{"codeSnapshot":"pass","durationMs":1,"results":[{"testId":"visible-case","status":"passed","message":"","durationMs":1},{"testId":"hidden-case","status":"passed","message":"","durationMs":1}]}`, http.StatusCreated},
		{http.MethodGet, "/api/courses/ai-ml/review-cards", "", http.StatusOK},
		{http.MethodPost, "/api/courses/ai-ml/modules/python/review-cards/functions.definition/reviews", `{"rating":"good"}`, http.StatusCreated},
		{http.MethodPost, "/api/tutor/turns", `{"message":"hello"}`, http.StatusOK},
		{http.MethodGet, "/api/tutor-access", "", http.StatusOK},
	}
	for _, request := range authenticatedRequests {
		response := authorizationRequest(t, app.Handler, request.method, request.path, request.body, cookie)
		if response.Code != request.status {
			t.Fatalf("authenticated %s %s = %d, want %d: %s", request.method, request.path, response.Code, request.status, response.Body.String())
		}
	}
}

func TestProtectedAPIUsesAuthenticatedOwner(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	catalog := testCatalog(t)
	learnerService := learner.NewService(db, catalog)
	authService := auth.NewService(db, auth.SessionConfig{TTL: time.Hour})
	if err := authService.ProvisionBootstrap(ctx, auth.BootstrapConfig{Username: "first", Password: "first-password", DisplayName: "First"}); err != nil {
		t.Fatal(err)
	}
	secondID := auth.UserID("00000000-0000-4000-8000-000000000002")
	hash, err := bcrypt.GenerateFromPassword([]byte("second-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (id, username, display_name, password_hash, created_at, updated_at) VALUES (?, 'second', 'Second', ?, '2026-08-23T00:00:00.000000000Z', '2026-08-23T00:00:00.000000000Z')`, secondID, hash); err != nil {
		t.Fatal(err)
	}
	_, firstToken, err := authService.SignIn(ctx, "first", "first-password")
	if err != nil {
		t.Fatal(err)
	}
	_, secondToken, err := authService.SignIn(ctx, "second", "second-password")
	if err != nil {
		t.Fatal(err)
	}
	app := NewAPIWithAuth(tutor.NewService(tutor.NewUnavailableProvider()), catalog, learnerService, review.NewService(db, catalog, review.SystemClock{}), authService)
	path := "/api/courses/ai-ml/modules/python/lessons/lesson.stable/progress"
	if response := authorizationRequest(t, app.Handler, http.MethodPut, path, `{"completed":true}`, authService.SessionCookie(firstToken)); response.Code != http.StatusOK {
		t.Fatalf("first learner update = %d: %s", response.Code, response.Body.String())
	}
	second := authorizationRequest(t, app.Handler, http.MethodGet, path, "", authService.SessionCookie(secondToken))
	if second.Code != http.StatusOK || !containsJSONField(second.Body.String(), `"completed":false`) {
		t.Fatalf("second learner observed first learner state: %d %s", second.Code, second.Body.String())
	}
	first := authorizationRequest(t, app.Handler, http.MethodGet, path+"?userId="+string(secondID), "", authService.SessionCookie(firstToken))
	if first.Code != http.StatusOK || !containsJSONField(first.Body.String(), `"completed":true`) {
		t.Fatalf("caller-controlled user identity changed owner selection: %d %s", first.Code, first.Body.String())
	}
}

func authorizationTestAPI(t *testing.T) (*API, *http.Cookie, *sql.DB) {
	t.Helper()
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open authorization state database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	catalog := testCatalog(t)
	learnerService := learner.NewService(db, catalog)
	reviewService := review.NewService(db, catalog, review.SystemClock{})
	app, cookie := newTestAPIWithSession(t, tutor.NewService(tutor.NewUnavailableProvider(), httpCostGate(t, true, 10)), catalog, learnerService, reviewService)
	return app, cookie, db
}

func authorizationRequest(t *testing.T, handler http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	return requestJSON(t, handler, method, path, body, cookie)
}

func allOperations(spec *huma.OpenAPI) map[string]*huma.Operation {
	result := make(map[string]*huma.Operation)
	for _, path := range spec.Paths {
		for _, operation := range []*huma.Operation{path.Get, path.Put, path.Post, path.Delete, path.Patch, path.Head, path.Options, path.Trace} {
			if operation != nil {
				result[operation.OperationID] = operation
			}
		}
	}
	return result
}

func containsJSONField(body, field string) bool { return strings.Contains(body, field) }
