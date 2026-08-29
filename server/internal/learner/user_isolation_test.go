package learner

import (
	"context"
	"database/sql"
	"testing"

	"github.com/johncrowleydev/helix-academy/server/internal/auth"
)

const secondLearnerUserID auth.UserID = "00000000-0000-4000-8000-000000000002"

func TestLearnerStateIsIsolatedByUser(t *testing.T) {
	service, db := evidenceTestService(t)
	insertLearnerTestUser(t, db, secondLearnerUserID)
	ctx := context.Background()

	if _, err := service.SetLessonProgress(ctx, testUserID, "course", "module", "lesson-one", true); err != nil {
		t.Fatalf("complete first user's lesson: %v", err)
	}
	secondProgress, err := service.LessonProgress(ctx, secondLearnerUserID, "course", "module", "lesson-one")
	if err != nil || secondProgress.Completed {
		t.Fatalf("second user observed first user's progress: %#v, %v", secondProgress, err)
	}
	if _, err := service.SetLessonProgress(ctx, secondLearnerUserID, "course", "module", "lesson-one", true); err != nil {
		t.Fatalf("complete second user's same lesson: %v", err)
	}

	if _, err := service.SetExerciseWorkspace(ctx, testUserID, "course", "module", "exercise-one", "first-code"); err != nil {
		t.Fatalf("save first workspace: %v", err)
	}
	if _, err := service.SetExerciseWorkspace(ctx, secondLearnerUserID, "course", "module", "exercise-one", "second-code"); err != nil {
		t.Fatalf("save second workspace: %v", err)
	}
	firstWorkspace, _ := service.ExerciseWorkspace(ctx, testUserID, "course", "module", "exercise-one")
	secondWorkspace, _ := service.ExerciseWorkspace(ctx, secondLearnerUserID, "course", "module", "exercise-one")
	if firstWorkspace.Code != "first-code" || secondWorkspace.Code != "second-code" {
		t.Fatalf("workspace ownership crossed users: first=%q second=%q", firstWorkspace.Code, secondWorkspace.Code)
	}

	results := []ExerciseTestResult{{TestID: "test", Status: "passed"}}
	if _, err := service.CreateExerciseAttempt(ctx, testUserID, "course", "module", "exercise-one", "first-code", 1, results); err != nil {
		t.Fatalf("create first user's attempt: %v", err)
	}
	secondAttempts, err := service.ExerciseAttempts(ctx, secondLearnerUserID, "course", "module", "exercise-one", 10)
	if err != nil || len(secondAttempts) != 0 {
		t.Fatalf("second user observed first user's attempts: %#v, %v", secondAttempts, err)
	}
	firstActivities, _ := service.Activities(ctx, testUserID, "course", 10)
	secondActivities, _ := service.Activities(ctx, secondLearnerUserID, "course", 10)
	if len(firstActivities) != 2 || len(secondActivities) != 1 {
		t.Fatalf("activities were not independently scoped: first=%d second=%d", len(firstActivities), len(secondActivities))
	}
}

func TestLearnerStateRejectsUnknownUser(t *testing.T) {
	service, _ := evidenceTestService(t)
	if _, err := service.SetLessonProgress(context.Background(), "00000000-0000-4000-8000-ffffffffffff", "course", "module", "lesson-one", true); err == nil {
		t.Fatal("expected unknown user ownership to violate the foreign key")
	}
}

func insertLearnerTestUser(t *testing.T, db *sql.DB, userID auth.UserID) {
	t.Helper()
	if _, err := db.Exec(`
		INSERT INTO users (id, display_name, created_at, updated_at)
		VALUES (?, 'Second learner', '2026-08-20T00:00:00.000000000Z', '2026-08-20T00:00:00.000000000Z')
	`, userID); err != nil {
		t.Fatalf("insert second learner: %v", err)
	}
}
