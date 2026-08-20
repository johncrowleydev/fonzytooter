package review

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v4"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
)

var testNow = time.Date(2026, time.August, 20, 14, 30, 0, 123456000, time.UTC)

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

func TestCardMappingRoundTripsEveryFSRSField(t *testing.T) {
	db, _ := newTestService(t, testNow)
	want := fsrs.Card{
		Due:            testNow.Add(72 * time.Hour),
		Stability:      18.75,
		Difficulty:     4.25,
		ScheduledDays:  3,
		Reps:           7,
		Lapses:         2,
		State:          fsrs.Relearning,
		LastReview:     testNow.Add(-24 * time.Hour),
		RemainingSteps: 1,
	}
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := storeCard(context.Background(), tx, "course", "module", "first", want, testNow); err != nil {
		t.Fatalf("store card: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	readTx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin read: %v", err)
	}
	got, found, err := loadCard(context.Background(), readTx, "course", "module", "first")
	_ = readTx.Rollback()
	if err != nil || !found {
		t.Fatalf("load card: found=%v err=%v", found, err)
	}
	assertCardsEqual(t, got, want)
}

func TestCardsTreatsAuthoredItemsAsVirtualNewWithoutWriting(t *testing.T) {
	db, service := newTestService(t, testNow)

	cards, err := service.Cards(context.Background(), "course", true)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("cards = %d, want 2", len(cards))
	}
	for _, card := range cards {
		if !card.Virtual || !card.New || !card.IsDue || len(card.Previews) != 4 {
			t.Fatalf("unexpected virtual card: %#v", card)
		}
		if !card.Schedule.Due.Equal(testNow) {
			t.Fatalf("virtual due = %v, want %v", card.Schedule.Due, testNow)
		}
	}
	assertTableCount(t, db, "review_cards", 0)
	assertTableCount(t, db, "review_logs", 0)
}

func TestAllRatingsUsePinnedSchedulerAndPreviewTimeBasis(t *testing.T) {
	for _, rating := range []Rating{RatingAgain, RatingHard, RatingGood, RatingEasy} {
		t.Run(string(rating), func(t *testing.T) {
			db, service := newTestService(t, testNow)
			cards, err := service.Cards(context.Background(), "course", true)
			if err != nil {
				t.Fatalf("preview: %v", err)
			}
			preview := previewFor(t, cards[0], rating)

			result, err := service.Submit(context.Background(), "course", "module", cards[0].Item.ID, rating)
			if err != nil {
				t.Fatalf("submit: %v", err)
			}
			grade, _ := fsrsRating(rating)
			want, err := fsrs.NewFSRS(fsrs.DefaultParam()).Next(fsrs.NewCard(testNow), testNow, grade)
			if err != nil {
				t.Fatalf("pinned scheduler: %v", err)
			}
			assertCardsEqual(t, result.Card.Schedule, want.Card)
			if !preview.DueAt.Equal(result.Card.Schedule.Due) {
				t.Fatalf("preview due %v != applied due %v", preview.DueAt, result.Card.Schedule.Due)
			}
			assertTableCount(t, db, "review_cards", 1)
			assertTableCount(t, db, "review_logs", 1)
			assertTableCountWhere(t, db, "activities", "kind = 'review_completed'", 1)

			var reviewedAt, dueAt, lastReviewAt string
			if err := db.QueryRow(`
				SELECT l.reviewed_at, c.due_at, c.last_review_at
				FROM review_logs l JOIN review_cards c
				ON c.course_id = l.course_id AND c.module_id = l.module_id
				AND c.review_item_id = l.review_item_id
			`).Scan(&reviewedAt, &dueAt, &lastReviewAt); err != nil {
				t.Fatalf("read UTC fields: %v", err)
			}
			for label, value := range map[string]string{"reviewed": reviewedAt, "due": dueAt, "last review": lastReviewAt} {
				parsed, err := time.Parse(time.RFC3339Nano, value)
				if err != nil || parsed.Location() != time.UTC {
					t.Fatalf("%s timestamp %q is not UTC: %v", label, value, err)
				}
			}
		})
	}
}

func TestDueOrderingPlacesStoredDueCardsBeforeNewCards(t *testing.T) {
	db, service := newTestService(t, testNow)
	card := fsrs.Card{
		Due: testNow.Add(-time.Hour), Stability: 2, Difficulty: 5,
		ScheduledDays: 1, Reps: 1, State: fsrs.Review,
		LastReview: testNow.Add(-24 * time.Hour),
	}
	tx, _ := db.BeginTx(context.Background(), nil)
	if err := storeCard(context.Background(), tx, "course", "module", "second", card, testNow); err != nil {
		t.Fatalf("store due card: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit due card: %v", err)
	}

	cards, err := service.Cards(context.Background(), "course", true)
	if err != nil {
		t.Fatalf("list due cards: %v", err)
	}
	if len(cards) != 2 || cards[0].Item.ID != "second" || cards[0].Virtual || cards[1].Item.ID != "first" || !cards[1].Virtual {
		t.Fatalf("unexpected due order: %#v", cards)
	}
}

func TestSubmitIsAtomicAcrossCardLogAndActivity(t *testing.T) {
	db, service := newTestService(t, testNow)
	if _, err := db.Exec("DROP TABLE activities"); err != nil {
		t.Fatalf("drop activities: %v", err)
	}

	if _, err := service.Submit(context.Background(), "course", "module", "first", RatingGood); err == nil {
		t.Fatal("expected activity insert failure")
	}
	assertTableCount(t, db, "review_cards", 0)
	assertTableCount(t, db, "review_logs", 0)
}

func TestInvalidAndMissingSubmissionsDoNotWrite(t *testing.T) {
	db, service := newTestService(t, testNow)
	for name, test := range map[string]struct {
		courseID string
		moduleID string
		itemID   string
		rating   Rating
		want     error
	}{
		"malformed rating": {"course", "module", "first", "perfect", ErrInvalidRating},
		"missing item":     {"course", "module", "missing", RatingGood, ErrReviewItemNotFound},
		"missing module":   {"course", "missing", "first", RatingGood, ErrReviewItemNotFound},
		"missing course":   {"missing", "module", "first", RatingGood, ErrCourseNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := service.Submit(context.Background(), test.courseID, test.moduleID, test.itemID, test.rating)
			if !errors.Is(err, test.want) {
				t.Fatalf("submit error = %v, want %v", err, test.want)
			}
		})
	}
	assertTableCount(t, db, "review_cards", 0)
	assertTableCount(t, db, "review_logs", 0)
}

func TestNoDueCardsAfterFutureScheduling(t *testing.T) {
	_, service := newSingleItemTestService(t, testNow)
	if _, err := service.Submit(context.Background(), "course", "module", "first", RatingEasy); err != nil {
		t.Fatalf("submit easy: %v", err)
	}
	cards, err := service.Cards(context.Background(), "course", true)
	if err != nil {
		t.Fatalf("list due: %v", err)
	}
	if len(cards) != 0 {
		t.Fatalf("due cards = %#v, want empty", cards)
	}
}

func newTestService(t *testing.T, now time.Time) (*sql.DB, *Service) {
	t.Helper()
	return newServiceWithReviewFiles(t, now, map[string]string{
		"first.yaml":  reviewYAML("first", 0),
		"second.yaml": reviewYAML("second", 1),
	})
}

func newSingleItemTestService(t *testing.T, now time.Time) (*sql.DB, *Service) {
	t.Helper()
	return newServiceWithReviewFiles(t, now, map[string]string{"first.yaml": reviewYAML("first", 0)})
}

func newServiceWithReviewFiles(t *testing.T, now time.Time, reviews map[string]string) (*sql.DB, *Service) {
	t.Helper()
	files := fstest.MapFS{
		"sources.yaml":                              {Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                {Data: []byte("id: course\ntitle: Course\ndescription: Test course.\norder: 0\n")},
		"courses/course/modules/module/module.yaml": {Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective\n    title: Objective\n    description: Test objective.\n    prerequisites: []\nvideos: []\nlessons:\n  - lesson\n")},
		"courses/course/modules/module/lesson.mdx":  {Data: []byte("---\nid: lesson\ntitle: Lesson\nobjectiveIds:\n  - objective\nsourceIds: []\n---\n# Lesson\n")},
	}
	for name, body := range reviews {
		files["courses/course/modules/module/reviews/"+name] = &fstest.MapFile{Data: []byte(body)}
	}
	catalog, err := curriculum.Load(files)
	if err != nil {
		t.Fatalf("load curriculum: %v", err)
	}
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "learner.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	completeLesson(t, db, "course", "module", "lesson", now)
	return db, NewService(db, catalog, fixedClock{now: now})
}

func TestCardsGateVirtualItemsOnSourceLessonButKeepPersistedSchedules(t *testing.T) {
	db, service := newTestService(t, testNow)
	if _, err := db.Exec("DELETE FROM lesson_progress"); err != nil {
		t.Fatalf("reset source lesson progress: %v", err)
	}

	cards, err := service.Cards(context.Background(), "course", true)
	if err != nil {
		t.Fatalf("list future cards: %v", err)
	}
	if len(cards) != 0 {
		t.Fatalf("future source lesson exposed virtual cards: %#v", cards)
	}
	if _, err := service.Submit(context.Background(), "course", "module", "first", RatingGood); !errors.Is(err, ErrReviewItemNotEligible) {
		t.Fatalf("submit future card error = %v, want %v", err, ErrReviewItemNotEligible)
	}

	persisted := fsrs.Card{
		Due: testNow.Add(-time.Hour), Stability: 2, Difficulty: 5,
		ScheduledDays: 1, Reps: 1, State: fsrs.Review,
		LastReview: testNow.Add(-24 * time.Hour),
	}
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin stored schedule: %v", err)
	}
	if err := storeCard(context.Background(), tx, "course", "module", "second", persisted, testNow); err != nil {
		t.Fatalf("store persisted schedule: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit persisted schedule: %v", err)
	}

	cards, err = service.Cards(context.Background(), "course", true)
	if err != nil {
		t.Fatalf("list persisted card: %v", err)
	}
	if len(cards) != 1 || cards[0].Item.ID != "second" || cards[0].Virtual {
		t.Fatalf("persisted schedule was gated with its source lesson: %#v", cards)
	}
}

func TestSubmitUsesModuleQualifiedReviewItemIdentity(t *testing.T) {
	files := fstest.MapFS{
		"sources.yaml":               {Data: []byte("sources: {}\n")},
		"courses/course/course.yaml": {Data: []byte("id: course\ntitle: Course\ndescription: Test course.\norder: 0\n")},
	}
	for index, moduleID := range []string{"first-module", "second-module"} {
		objectiveID := moduleID + ".objective"
		moduleRoot := "courses/course/modules/" + moduleID + "/"
		files[moduleRoot+"module.yaml"] = &fstest.MapFile{Data: []byte(fmt.Sprintf("id: %s\ntitle: %s\norder: %d\nobjectives:\n  - id: %s\n    title: Objective\n    description: Test objective.\n    prerequisites: []\nvideos: []\nlessons:\n  - lesson\n", moduleID, moduleID, index, objectiveID))}
		files[moduleRoot+"lesson.mdx"] = &fstest.MapFile{Data: []byte(fmt.Sprintf("---\nid: lesson\ntitle: Lesson\nobjectiveIds:\n  - %s\nsourceIds: []\n---\n# Lesson\n", objectiveID))}
		files[moduleRoot+"reviews/shared.yaml"] = &fstest.MapFile{Data: []byte(fmt.Sprintf("id: shared\norder: 0\nobjectiveIds:\n  - %s\nsourceLessonId: lesson\nprompt: Prompt?\nanswer: Answer.\n", objectiveID))}
	}
	catalog, err := curriculum.Load(files)
	if err != nil {
		t.Fatalf("load duplicate review ids: %v", err)
	}
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "learner.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	completeLesson(t, db, "course", "second-module", "lesson", testNow)
	service := NewService(db, catalog, fixedClock{now: testNow})

	if _, err := service.Submit(context.Background(), "course", "second-module", "shared", RatingGood); err != nil {
		t.Fatalf("submit module-qualified review: %v", err)
	}
	var moduleID string
	if err := db.QueryRow("SELECT module_id FROM review_cards WHERE review_item_id = 'shared'").Scan(&moduleID); err != nil {
		t.Fatalf("read stored review identity: %v", err)
	}
	if moduleID != "second-module" {
		t.Fatalf("stored review module = %q, want second-module", moduleID)
	}
}

func completeLesson(t *testing.T, db *sql.DB, courseID, moduleID, lessonID string, now time.Time) {
	t.Helper()
	if _, err := db.Exec(`
		INSERT INTO lesson_progress (course_id, module_id, lesson_id, completed, completed_at, updated_at)
		VALUES (?, ?, ?, 1, ?, ?)
	`, courseID, moduleID, lessonID, formatTime(now), formatTime(now)); err != nil {
		t.Fatalf("complete source lesson: %v", err)
	}
}

func reviewYAML(id string, order int) string {
	return fmt.Sprintf("id: %s\norder: %d\n", id, order) +
		"objectiveIds:\n  - objective\n" +
		"sourceLessonId: lesson\n" +
		"prompt: What is the neutral answer?\n" +
		"answer: The neutral answer.\n" +
		"hint: Recall the neutral fixture.\n"
}

func previewFor(t *testing.T, card Card, rating Rating) Preview {
	t.Helper()
	for _, preview := range card.Previews {
		if preview.Rating == rating {
			return preview
		}
	}
	t.Fatalf("missing preview %q", rating)
	return Preview{}
}

func assertCardsEqual(t *testing.T, got, want fsrs.Card) {
	t.Helper()
	if !got.Due.Equal(want.Due) || !got.LastReview.Equal(want.LastReview) ||
		math.Abs(got.Stability-want.Stability) > 1e-12 ||
		math.Abs(got.Difficulty-want.Difficulty) > 1e-12 ||
		got.ScheduledDays != want.ScheduledDays || got.Reps != want.Reps ||
		got.Lapses != want.Lapses || got.State != want.State ||
		got.RemainingSteps != want.RemainingSteps {
		t.Fatalf("card mismatch:\n got %#v\nwant %#v", got, want)
	}
}

func assertTableCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	assertTableCountWhere(t, db, table, "1 = 1", want)
}

func assertTableCountWhere(t *testing.T, db *sql.DB, table, where string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table + " WHERE " + where).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("%s count = %d, want %d", table, got, want)
	}
}
