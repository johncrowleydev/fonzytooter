package learner

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"testing/fstest"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
)

func TestCourseProgressAggregatesReviewAndExerciseEvidence(t *testing.T) {
	service, db := evidenceTestService(t)
	now := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	insertEvidenceRow(t, db, `INSERT INTO lesson_progress VALUES (?, ?, ?, 1, ?, ?)`, "course", "module", "lesson-one", timestamp(now.Add(-2*time.Hour)), timestamp(now.Add(-2*time.Hour)))
	insertReviewCard(t, db, "review-future", now.Add(24*time.Hour), now.Add(-time.Hour))
	insertEvidenceRow(t, db, `
		INSERT INTO review_logs (
			course_id, module_id, review_item_id, reviewed_at, rating, previous_due, next_due,
			before_stability, after_stability, before_difficulty, after_difficulty,
			before_scheduled_days, after_scheduled_days, before_reps, after_reps,
			before_lapses, after_lapses, before_state, after_state,
			before_last_review_at, after_last_review_at, before_remaining_steps, after_remaining_steps
		) VALUES ('course', 'module', 'review-future', ?, 'good', ?, ?, 0, 1, 0, 5, 0, 1, 0, 1, 0, 0, 0, 1, NULL, ?, 0, 0)
	`, timestamp(now.Add(-time.Hour)), timestamp(now.Add(-time.Hour)), timestamp(now.Add(24*time.Hour)), timestamp(now.Add(-time.Hour)))
	insertAttempt(t, db, "exercise-one", now.Add(-30*time.Minute), false)
	insertAttempt(t, db, "exercise-two", now.Add(-15*time.Minute), true)

	progress, err := service.CourseProgress(context.Background(), "course")
	if err != nil {
		t.Fatalf("course progress: %v", err)
	}
	if progress.DueReviewCount != 1 || progress.NextLesson == nil || progress.NextLesson.LessonID != "lesson-two" {
		t.Fatalf("unexpected course-level state: %#v", progress)
	}
	if progress.PracticeExercise == nil || progress.PracticeExercise.ExerciseID != "exercise-one" {
		t.Fatalf("expected first introduced exercise without a passing attempt, got %#v", progress.PracticeExercise)
	}
	objective := progress.Objectives[0]
	if !objective.Introduced || objective.LinkedLessonCount != 2 || objective.CompletedLessonCount != 1 {
		t.Fatalf("unexpected introduction evidence: %#v", objective)
	}
	if objective.Recall.ReviewItemCount != 2 || objective.Recall.ReviewsCompleted != 1 || objective.Recall.DueReviewCount != 1 || objective.Recall.LastReviewedAt == nil || objective.Recall.NextDueAt == nil || !objective.Recall.NextDueAt.Equal(now) {
		t.Fatalf("unexpected recall evidence: %#v", objective.Recall)
	}
	if objective.Application.ExerciseCount != 2 || objective.Application.Attempts != 2 || objective.Application.FullyPassedAttempts != 1 || objective.Application.LastCheckedAt == nil || !objective.Application.LastCheckedAt.Equal(now.Add(-15*time.Minute)) {
		t.Fatalf("unexpected application evidence: %#v", objective.Application)
	}
	if objective.TransferAssessed {
		t.Fatal("transfer must remain unassessed")
	}
}

func TestCourseProgressUsesRecentCompletionThenFallsBackToFirstIncomplete(t *testing.T) {
	service, db := evidenceTestService(t)
	now := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	insertEvidenceRow(t, db, `INSERT INTO lesson_progress VALUES (?, ?, ?, 1, ?, ?)`, "course", "module", "lesson-one", timestamp(now.Add(-2*time.Hour)), timestamp(now.Add(-2*time.Hour)))

	progress, err := service.CourseProgress(context.Background(), "course")
	if err != nil || progress.NextLesson == nil || progress.NextLesson.LessonID != "lesson-two" {
		t.Fatalf("expected lesson after most recent completion, got %#v, %v", progress.NextLesson, err)
	}
	insertEvidenceRow(t, db, `INSERT INTO lesson_progress VALUES (?, ?, ?, 1, ?, ?)`, "course", "module", "lesson-three", timestamp(now), timestamp(now))
	progress, err = service.CourseProgress(context.Background(), "course")
	if err != nil || progress.NextLesson == nil || progress.NextLesson.LessonID != "lesson-two" {
		t.Fatalf("expected first incomplete fallback, got %#v, %v", progress.NextLesson, err)
	}
	insertEvidenceRow(t, db, `INSERT INTO lesson_progress VALUES (?, ?, ?, 1, ?, ?)`, "course", "module", "lesson-two", timestamp(now.Add(time.Hour)), timestamp(now.Add(time.Hour)))
	progress, err = service.CourseProgress(context.Background(), "course")
	if err != nil || progress.NextLesson != nil {
		t.Fatalf("expected all-lessons-complete state, got %#v, %v", progress.NextLesson, err)
	}
}

func evidenceTestService(t *testing.T) (*Service, *sql.DB) {
	t.Helper()
	catalog, err := curriculum.Load(fstest.MapFS{
		"sources.yaml":                                     &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                       &fstest.MapFile{Data: []byte("id: course\ntitle: Course\ndescription: Test course.\norder: 0\n")},
		"courses/course/modules/module/module.yaml":        &fstest.MapFile{Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective.one\n    title: Objective one\n    description: First objective.\n    prerequisites: []\n  - id: objective.two\n    title: Objective two\n    description: Second objective.\n    prerequisites: []\nvideos: []\nlessons:\n  - lesson-one\n  - lesson-two\n  - lesson-three\n")},
		"courses/course/modules/module/lesson-one.mdx":     &fstest.MapFile{Data: []byte("---\nid: lesson-one\ntitle: Lesson one\nobjectiveIds: [objective.one]\nsourceIds: []\n---\nOne\n")},
		"courses/course/modules/module/lesson-two.mdx":     &fstest.MapFile{Data: []byte("---\nid: lesson-two\ntitle: Lesson two\nobjectiveIds: [objective.one]\nsourceIds: []\n---\nTwo\n")},
		"courses/course/modules/module/lesson-three.mdx":   &fstest.MapFile{Data: []byte("---\nid: lesson-three\ntitle: Lesson three\nobjectiveIds: [objective.two]\nsourceIds: []\n---\nThree\n")},
		"courses/course/modules/module/exercises/one.yaml": &fstest.MapFile{Data: []byte(exerciseFixture("exercise-one", 0))},
		"courses/course/modules/module/exercises/two.yaml": &fstest.MapFile{Data: []byte(exerciseFixture("exercise-two", 1))},
		"courses/course/modules/module/reviews/one.yaml":   &fstest.MapFile{Data: []byte(reviewFixture("review-due", 0))},
		"courses/course/modules/module/reviews/two.yaml":   &fstest.MapFile{Data: []byte(reviewFixture("review-future", 1))},
	})
	if err != nil {
		t.Fatalf("load evidence catalog: %v", err)
	}
	db, err := database.Open(context.Background(), t.TempDir()+"/evidence.db")
	if err != nil {
		t.Fatalf("open evidence database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewService(db, catalog), db
}

func exerciseFixture(id string, order int) string {
	return fmt.Sprintf("id: %s\ntitle: %s\nlessonId: lesson-one\norder: %d\nobjectiveIds: [objective.one]\nprompt: Implement it.\nstarterCode: pass\ntests:\n  - id: test\n    title: Test\n    visibility: visible\n    code: assert True\n", id, id, order)
}

func reviewFixture(id string, order int) string {
	return fmt.Sprintf("id: %s\norder: %d\nobjectiveIds: [objective.one]\nsourceLessonId: lesson-one\nprompt: Prompt?\nanswer: Answer.\n", id, order)
}

func insertReviewCard(t *testing.T, db *sql.DB, itemID string, dueAt, reviewedAt time.Time) {
	t.Helper()
	insertEvidenceRow(t, db, `
		INSERT INTO review_cards (
			course_id, module_id, review_item_id, due_at, stability, difficulty,
			scheduled_days, reps, lapses, state, last_review_at, remaining_steps, updated_at
		) VALUES ('course', 'module', ?, ?, 1, 5, 1, 1, 0, 1, ?, 0, ?)
	`, itemID, timestamp(dueAt), timestamp(reviewedAt), timestamp(reviewedAt))
}

func insertAttempt(t *testing.T, db *sql.DB, exerciseID string, createdAt time.Time, allPassed bool) {
	t.Helper()
	insertEvidenceRow(t, db, `
		INSERT INTO exercise_attempts (
			course_id, module_id, exercise_id, created_at, passed_count, failed_count,
			duration_ms, all_passed, code_snapshot
		) VALUES ('course', 'module', ?, ?, ?, ?, 10, ?, 'pass')
	`, exerciseID, timestamp(createdAt), boolInt(allPassed), boolInt(!allPassed), boolInt(allPassed))
}

func insertEvidenceRow(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("insert evidence row: %v", err)
	}
}

func timestamp(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
