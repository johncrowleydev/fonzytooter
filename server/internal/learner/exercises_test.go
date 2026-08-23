package learner

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
)

func TestExerciseWorkspaceDefaultsAndPersists(t *testing.T) {
	service, db := exerciseTestService(t)
	workspace, err := service.ExerciseWorkspace(context.Background(), testUserID, "course", "module", "double")
	if err != nil {
		t.Fatalf("get default workspace: %v", err)
	}
	if workspace.Code != "def double(x):\n    pass\n" || workspace.UpdatedAt != nil {
		t.Fatalf("unexpected default workspace: %#v", workspace)
	}
	now := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	saved, err := service.SetExerciseWorkspace(context.Background(), testUserID, "course", "module", "double", "def double(x): return x * 2")
	if err != nil {
		t.Fatalf("save workspace: %v", err)
	}
	if saved.UpdatedAt == nil || !saved.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected saved workspace: %#v", saved)
	}
	if _, err := service.SetExerciseWorkspace(context.Background(), testUserID, "course", "module", "double", "updated"); err != nil {
		t.Fatalf("replace workspace: %v", err)
	}
	assertRowCount(t, db, "exercise_workspaces", 1)
	if _, err := service.ExerciseWorkspace(context.Background(), testUserID, "course", "wrong", "double"); !errors.Is(err, ErrExerciseNotFound) {
		t.Fatalf("expected ownership validation, got %v", err)
	}
}

func TestCreateExerciseAttemptDerivesAggregatesAndActivity(t *testing.T) {
	service, db := exerciseTestService(t)
	now := time.Date(2026, time.August, 20, 12, 30, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	results := []ExerciseTestResult{
		{TestID: "hidden", Status: "failed", Message: "expected 0", DurationMS: 5},
		{TestID: "visible", Status: "passed", DurationMS: 4},
	}
	attempt, err := service.CreateExerciseAttempt(context.Background(), testUserID, "course", "module", "double", "code", 14, results)
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	if attempt.PassedCount != 1 || attempt.FailedCount != 1 || attempt.AllPassed || !attempt.CreatedAt.Equal(now) {
		t.Fatalf("server did not derive aggregates: %#v", attempt)
	}
	if attempt.Results[0].TestID != "visible" || attempt.Results[1].TestID != "hidden" {
		t.Fatalf("attempt results were not normalized to authored order: %#v", attempt.Results)
	}
	assertRowCount(t, db, "exercise_attempts", 1)
	assertRowCount(t, db, "exercise_test_results", 2)
	assertRowCount(t, db, "activities", 1)
	var kind string
	if err := db.QueryRow("SELECT kind FROM activities").Scan(&kind); err != nil || kind != ActivityExerciseChecked {
		t.Fatalf("unexpected activity: %q, %v", kind, err)
	}
}

func TestRecentExerciseAttemptsAndActivitiesUseChronologicalFixedWidthTimestamps(t *testing.T) {
	service, db := exerciseTestService(t)
	base := time.Date(2026, time.August, 20, 12, 30, 0, 0, time.UTC)
	times := []time.Time{base.Add(90 * time.Millisecond), base.Add(100 * time.Millisecond), base.Add(110 * time.Millisecond)}
	index := 0
	service.now = func() time.Time {
		value := times[index]
		index++
		return value
	}
	results := []ExerciseTestResult{
		{TestID: "visible", Status: "passed"},
		{TestID: "hidden", Status: "passed"},
	}
	created := make([]ExerciseAttempt, 0, len(times))
	for range times {
		attempt, err := service.CreateExerciseAttempt(context.Background(), testUserID, "course", "module", "double", "code", 1, results)
		if err != nil {
			t.Fatalf("create exercise attempt: %v", err)
		}
		created = append(created, attempt)
	}
	var stored string
	if err := db.QueryRow(`SELECT created_at FROM exercise_attempts WHERE id = ?`, created[1].ID).Scan(&stored); err != nil {
		t.Fatalf("read fixed-width exercise timestamp: %v", err)
	}
	if stored != "2026-08-20T12:30:00.100000000Z" {
		t.Fatalf("expected fixed-width exercise timestamp, got %q", stored)
	}
	attempts, err := service.ExerciseAttempts(context.Background(), testUserID, "course", "module", "double", 1)
	if err != nil {
		t.Fatalf("read recent exercise attempts: %v", err)
	}
	if len(attempts) != 1 || attempts[0].ID != created[2].ID {
		t.Fatalf("expected 110ms exercise attempt, got %#v", attempts)
	}
	activities, err := service.Activities(context.Background(), testUserID, "course", 1)
	if err != nil {
		t.Fatalf("read recent activities: %v", err)
	}
	if len(activities) != 1 || !activities[0].OccurredAt.Equal(times[2]) {
		t.Fatalf("expected 110ms activity, got %#v", activities)
	}
}

func TestCreateExerciseAttemptRejectsUnknownDuplicateAndIncompleteTests(t *testing.T) {
	service, db := exerciseTestService(t)
	tests := []struct {
		name    string
		results []ExerciseTestResult
	}{
		{name: "unknown", results: []ExerciseTestResult{{TestID: "visible", Status: "passed"}, {TestID: "unknown", Status: "passed"}}},
		{name: "duplicate", results: []ExerciseTestResult{{TestID: "visible", Status: "passed"}, {TestID: "visible", Status: "passed"}}},
		{name: "incomplete", results: []ExerciseTestResult{{TestID: "visible", Status: "passed"}}},
		{name: "invalid status", results: []ExerciseTestResult{{TestID: "visible", Status: "maybe"}, {TestID: "hidden", Status: "passed"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := service.CreateExerciseAttempt(context.Background(), testUserID, "course", "module", "double", "code", 1, test.results); !errors.Is(err, ErrInvalidExerciseAttempt) {
				t.Fatalf("expected invalid attempt, got %v", err)
			}
		})
	}
	assertRowCount(t, db, "exercise_attempts", 0)
	assertRowCount(t, db, "activities", 0)
}

func exerciseTestService(t *testing.T) (*Service, *sql.DB) {
	t.Helper()
	catalog := exerciseCatalog(t)
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "fonzytooter.db"))
	if err != nil {
		t.Fatalf("open exercise test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewService(db, catalog), db
}

func exerciseCatalog(t *testing.T) *curriculum.Catalog {
	t.Helper()
	fsys := fstest.MapFS{
		"sources.yaml":                                        {Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                          {Data: []byte("id: course\ntitle: Course\ndescription: Test.\norder: 0\n")},
		"courses/course/modules/module/module.yaml":           {Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective\n    title: Objective\n    description: Test.\n    prerequisites: []\nvideos: []\nlessons:\n  - lesson\n")},
		"courses/course/modules/module/lesson.mdx":            {Data: []byte("---\nid: lesson\ntitle: Lesson\nobjectiveIds:\n  - objective\nsourceIds: []\n---\n# Lesson\n")},
		"courses/course/modules/module/exercises/double.yaml": {Data: []byte("id: double\ntitle: Double\nlessonId: lesson\norder: 0\nobjectiveIds:\n  - objective\nprompt: Double it.\nstarterCode: |\n  def double(x):\n      pass\ntests:\n  - id: visible\n    title: Doubles two\n    visibility: visible\n    code: assert double(2) == 4\n  - id: hidden\n    title: Doubles zero\n    visibility: hidden\n    code: assert double(0) == 0\n")},
	}
	catalog, err := curriculum.Load(fsys)
	if err != nil {
		t.Fatalf("load exercise catalog: %v", err)
	}
	return catalog
}
