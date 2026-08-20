package learner

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
)

func TestLessonProgressLifecycleAndActivityIdempotency(t *testing.T) {
	service, db, catalog := testService(t)
	course, module, lesson := firstLesson(t, catalog)
	testNow := time.Date(2026, time.August, 20, 4, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return testNow }

	progress, err := service.LessonProgress(context.Background(), course.ID, module.ID, lesson.ID)
	if err != nil {
		t.Fatalf("get default progress: %v", err)
	}
	if progress.Completed || progress.CompletedAt != nil {
		t.Fatalf("expected incomplete default progress, got %#v", progress)
	}

	completed, err := service.SetLessonProgress(context.Background(), course.ID, module.ID, lesson.ID, true)
	if err != nil {
		t.Fatalf("complete lesson: %v", err)
	}
	if !completed.Completed || completed.CompletedAt == nil || !completed.CompletedAt.Equal(testNow) {
		t.Fatalf("unexpected completed progress: %#v", completed)
	}
	if _, err := service.SetLessonProgress(context.Background(), course.ID, module.ID, lesson.ID, true); err != nil {
		t.Fatalf("repeat completion: %v", err)
	}

	assertRowCount(t, db, "lesson_progress", 1)
	assertRowCount(t, db, "activities", 1)

	incomplete, err := service.SetLessonProgress(context.Background(), course.ID, module.ID, lesson.ID, false)
	if err != nil {
		t.Fatalf("uncomplete lesson: %v", err)
	}
	if incomplete.Completed || incomplete.CompletedAt != nil {
		t.Fatalf("unexpected incomplete progress: %#v", incomplete)
	}
	assertRowCount(t, db, "lesson_progress", 1)
	assertRowCount(t, db, "activities", 1)
}

func TestLessonProgressValidatesCatalogOwnership(t *testing.T) {
	service, _, catalog := testService(t)
	course, module, lesson := firstLesson(t, catalog)

	tests := []struct {
		name     string
		courseID string
		moduleID string
		lessonID string
	}{
		{name: "missing course", courseID: "missing", moduleID: module.ID, lessonID: lesson.ID},
		{name: "missing module", courseID: course.ID, moduleID: "missing", lessonID: lesson.ID},
		{name: "missing lesson", courseID: course.ID, moduleID: module.ID, lessonID: "missing"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := service.LessonProgress(context.Background(), test.courseID, test.moduleID, test.lessonID); !errors.Is(err, ErrLessonNotFound) {
				t.Fatalf("expected lesson not found, got %v", err)
			}
			if _, err := service.SetLessonProgress(context.Background(), test.courseID, test.moduleID, test.lessonID, true); !errors.Is(err, ErrLessonNotFound) {
				t.Fatalf("expected lesson not found on update, got %v", err)
			}
		})
	}
}

func TestCourseProgressDerivesIntroducedObjectivesAndNextLesson(t *testing.T) {
	service, _, catalog := testService(t)
	course := catalog.Courses()[0]
	lessons := courseLessons(course)
	if len(lessons) < 2 {
		t.Fatal("test curriculum needs at least two lessons")
	}

	initial, err := service.CourseProgress(context.Background(), course.ID)
	if err != nil {
		t.Fatalf("get initial course progress: %v", err)
	}
	if initial.NextLesson == nil || initial.NextLesson.LessonID != lessons[0].lesson.ID {
		t.Fatalf("expected first lesson next, got %#v", initial.NextLesson)
	}
	if initial.CompletedLessonCount != 0 || initial.TotalLessonCount != len(lessons) {
		t.Fatalf("unexpected initial counts: %#v", initial)
	}

	first := lessons[0]
	if _, err := service.SetLessonProgress(context.Background(), course.ID, first.module.ID, first.lesson.ID, true); err != nil {
		t.Fatalf("complete first lesson: %v", err)
	}
	progress, err := service.CourseProgress(context.Background(), course.ID)
	if err != nil {
		t.Fatalf("get course progress: %v", err)
	}
	if progress.NextLesson == nil || progress.NextLesson.LessonID != lessons[1].lesson.ID {
		t.Fatalf("expected second lesson next, got %#v", progress.NextLesson)
	}
	introduced := make(map[string]bool, len(progress.Objectives))
	for _, objective := range progress.Objectives {
		introduced[objective.ID] = objective.Introduced
		if objective.Recall != EvidenceNotAssessed || objective.Application != EvidenceNotAssessed || objective.Transfer != EvidenceNotAssessed {
			t.Fatalf("fabricated evidence dimensions for %s: %#v", objective.ID, objective)
		}
	}
	for _, objectiveID := range first.lesson.ObjectiveIDs {
		if !introduced[objectiveID] {
			t.Fatalf("expected objective %s to be introduced", objectiveID)
		}
	}

	for _, item := range lessons[1:] {
		if _, err := service.SetLessonProgress(context.Background(), course.ID, item.module.ID, item.lesson.ID, true); err != nil {
			t.Fatalf("complete lesson %s: %v", item.lesson.ID, err)
		}
	}
	complete, err := service.CourseProgress(context.Background(), course.ID)
	if err != nil {
		t.Fatalf("get completed course progress: %v", err)
	}
	if complete.NextLesson != nil || complete.CompletedLessonCount != complete.TotalLessonCount {
		t.Fatalf("expected completed course, got %#v", complete)
	}
	if _, err := service.CourseProgress(context.Background(), "missing"); !errors.Is(err, ErrCourseNotFound) {
		t.Fatalf("expected missing course error, got %v", err)
	}
}

func TestActivitiesAreNewestFirstBoundedAndEnriched(t *testing.T) {
	service, _, catalog := testService(t)
	course := catalog.Courses()[0]
	lessons := courseLessons(course)
	if len(lessons) < 2 {
		t.Fatal("test curriculum needs at least two lessons")
	}
	testNow := time.Date(2026, time.August, 20, 5, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return testNow }

	for _, item := range lessons[:2] {
		if _, err := service.SetLessonProgress(context.Background(), course.ID, item.module.ID, item.lesson.ID, true); err != nil {
			t.Fatalf("complete lesson %s: %v", item.lesson.ID, err)
		}
	}
	activities, err := service.Activities(context.Background(), course.ID, 1)
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(activities) != 1 || activities[0].LessonID == nil || *activities[0].LessonID != lessons[1].lesson.ID {
		t.Fatalf("expected newest activity with deterministic ID tie-break, got %#v", activities)
	}
	if activities[0].CourseTitle != course.Title || activities[0].ModuleTitle == nil || activities[0].LessonTitle == nil {
		t.Fatalf("expected curriculum-enriched titles, got %#v", activities[0])
	}
	empty, err := service.Activities(context.Background(), catalog.Courses()[0].ID, DefaultActivityLimit)
	if err != nil || len(empty) != 2 {
		t.Fatalf("expected two activities, got %d, %v", len(empty), err)
	}
	if _, err := service.Activities(context.Background(), "missing", 20); !errors.Is(err, ErrCourseNotFound) {
		t.Fatalf("expected missing course error, got %v", err)
	}
}

type courseLesson struct {
	module curriculum.Module
	lesson curriculum.Lesson
}

func testService(t *testing.T) (*Service, *sql.DB, *curriculum.Catalog) {
	t.Helper()
	catalog, err := curriculum.Load(os.DirFS(filepath.Join("..", "..", "..", "curriculum")))
	if err != nil {
		t.Fatalf("load test curriculum: %v", err)
	}
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "fonzytooter.db"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewService(db, catalog), db, catalog
}

func firstLesson(t *testing.T, catalog *curriculum.Catalog) (curriculum.Course, curriculum.Module, curriculum.Lesson) {
	t.Helper()
	course := catalog.Courses()[0]
	module := course.Modules[0]
	if len(module.Lessons) == 0 {
		t.Fatal("test curriculum needs a lesson in the first module")
	}
	return course, module, module.Lessons[0]
}

func courseLessons(course curriculum.Course) []courseLesson {
	lessons := make([]courseLesson, 0)
	for _, module := range course.Modules {
		for _, lesson := range module.Lessons {
			lessons = append(lessons, courseLesson{module: module, lesson: lesson})
		}
	}
	return lessons
}

func assertRowCount(t *testing.T, db *sql.DB, table string, expected int) {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if count != expected {
		t.Fatalf("expected %d %s rows, got %d", expected, table, count)
	}
}
