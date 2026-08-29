package curriculumstate

import (
	"bytes"
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
	"github.com/johncrowleydev/helix-academy/server/internal/curriculumidentity"
	"github.com/johncrowleydev/helix-academy/server/internal/database"
)

func TestAuditCleanDatabaseHasNoFalsePositives(t *testing.T) {
	db := openTestDatabase(t)
	seedHistory(t, db, identitySet{"lesson", "exercise", "test", "review"})

	report, err := Audit(context.Background(), db, catalog(t, identitySet{"lesson", "exercise", "test", "review"}))
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	if !report.Clean() {
		t.Fatalf("expected clean report, got %#v", report.Findings)
	}
	var output bytes.Buffer
	if err := WriteReport(&output, report); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "curriculum state valid: no orphaned curriculum references\n" {
		t.Fatalf("unexpected clean output %q", got)
	}
}

func TestAuditReportsRemovedLessonExerciseTestAndReviewHistoryDeterministically(t *testing.T) {
	db := openTestDatabase(t)
	seedHistory(t, db, identitySet{"lesson", "exercise", "test", "review"})
	current := catalog(t, identitySet{"lesson-new", "exercise-new", "test-new", "review-new"})

	report, err := Audit(context.Background(), db, current)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	var output bytes.Buffer
	if err := WriteReport(&output, report); err != nil {
		t.Fatal(err)
	}
	got := output.String()
	wantGroups := []string{
		"activities (3)",
		"exercise_attempts (1)",
		"exercise_test_results (1)",
		"exercise_workspaces (1)",
		"lesson_progress (1)",
		"review_cards (1)",
		"review_logs (1)",
	}
	last := -1
	for _, group := range wantGroups {
		index := strings.Index(got, group)
		if index <= last {
			t.Fatalf("expected deterministic group %q after index %d:\n%s", group, last, got)
		}
		last = index
	}
	for _, detail := range []string{
		"ai-ml/module/lesson",
		"ai-ml/module/exercise/test",
		"ai-ml/module/review",
		"id=1 kind=lesson_completed",
	} {
		if !strings.Contains(got, detail) {
			t.Fatalf("expected identifying detail %q:\n%s", detail, got)
		}
	}
}

func TestAuditValidatesKnownActivityIdentityColumns(t *testing.T) {
	db := openTestDatabase(t)
	if _, err := db.Exec(`INSERT INTO activities (user_id, kind, course_id, module_id, occurred_at) VALUES ('00000000-0000-4000-8000-000000000001', 'exercise_checked', 'ai-ml', 'module', '2026-01-01')`); err != nil {
		t.Fatal(err)
	}
	report, err := Audit(context.Background(), db, catalog(t, identitySet{"lesson", "exercise", "test", "review"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Findings) != 1 || !strings.Contains(report.Findings[0].Reason, "has no exercise_id") {
		t.Fatalf("expected malformed activity finding, got %#v", report.Findings)
	}
}

func TestApplyMigrationsPreservesLessonExerciseAndReviewHistory(t *testing.T) {
	db := openTestDatabase(t)
	seedHistory(t, db, identitySet{"lesson", "exercise", "test", "review"})
	migrations := parseMigrations(t, `version: 1
migrations:
  - entity: lesson
    from: ai-ml/module/lesson
    to: ai-ml/module/lesson-new
  - entity: exercise
    from: ai-ml/module/exercise
    to: ai-ml/module/exercise-new
  - entity: exercise-test
    from: ai-ml/module/exercise/test
    to: ai-ml/module/exercise-new/test-new
  - entity: review-item
    from: ai-ml/module/review
    to: ai-ml/module/review-new
`)

	current := catalog(t, identitySet{"lesson-new", "exercise-new", "test-new", "review-new"})
	updates, err := ApplyMigrations(context.Background(), db, current, migrations)
	if err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if len(updates) != 4 {
		t.Fatalf("expected four migration results, got %#v", updates)
	}
	report, err := Audit(context.Background(), db, current)
	if err != nil {
		t.Fatalf("audit migrated database: %v", err)
	}
	if !report.Clean() {
		t.Fatalf("expected migrated state to resolve, got %#v", report.Findings)
	}

	assertCount(t, db, `SELECT COUNT(*) FROM lesson_progress WHERE lesson_id = 'lesson-new' AND completed = 1`, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM exercise_workspaces WHERE exercise_id = 'exercise-new' AND code = 'saved code'`, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM exercise_attempts WHERE exercise_id = 'exercise-new' AND code_snapshot = 'attempt code'`, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM exercise_test_results WHERE test_id = 'test-new' AND message = 'history'`, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM review_cards WHERE review_item_id = 'review-new' AND reps = 2`, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM review_logs WHERE review_item_id = 'review-new' AND rating = 'good'`, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM activities WHERE lesson_id = 'lesson-new' OR exercise_id = 'exercise-new' OR review_item_id = 'review-new'`, 3)
	again, err := ApplyMigrations(context.Background(), db, current, migrations)
	if err != nil {
		t.Fatalf("reapply migrations: %v", err)
	}
	for _, update := range again {
		if update.Updates != 0 {
			t.Fatalf("expected idempotent migration, got %#v", again)
		}
	}
}

func TestVideoStateAuditAndMigrationPreserveStableIdentity(t *testing.T) {
	db := openTestDatabase(t)
	if _, err := db.Exec(`INSERT INTO video_progress (user_id, course_id, module_id, video_id, completed, completed_at, updated_at) VALUES ('00000000-0000-4000-8000-000000000001', 'course', 'module', 'old-video', 1, '2026-01-01', '2026-01-01')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO activities (user_id, kind, course_id, module_id, video_id, occurred_at) VALUES ('00000000-0000-4000-8000-000000000001', 'video_completed', 'course', 'module', 'old-video', '2026-01-01')`); err != nil {
		t.Fatal(err)
	}
	current := videoCatalog(t, "new-video")
	report, err := Audit(context.Background(), db, current)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Findings) != 2 {
		t.Fatalf("expected retired video state findings, got %#v", report.Findings)
	}

	migrations := parseMigrations(t, `version: 1
migrations:
  - entity: video
    from: course/module/old-video
    to: course/module/new-video
`)
	updates, err := ApplyMigrations(context.Background(), db, current, migrations)
	if err != nil {
		t.Fatalf("migrate video identity: %v", err)
	}
	if len(updates) != 1 || updates[0].Updates != 2 {
		t.Fatalf("unexpected video migration updates: %#v", updates)
	}
	report, err = Audit(context.Background(), db, current)
	if err != nil || !report.Clean() {
		t.Fatalf("migrated video state did not resolve: %#v, %v", report.Findings, err)
	}
}

func videoCatalog(t *testing.T, videoID string) *curriculum.Catalog {
	t.Helper()
	catalog, err := curriculum.Load(fstest.MapFS{
		"sources.yaml":                              {Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                {Data: []byte("id: course\ntitle: Course\ndescription: Course.\norder: 0\n")},
		"courses/course/modules/module/module.yaml": {Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective\n    title: Objective\n    description: Objective.\n    prerequisites: []\nvideos:\n  - id: " + videoID + "\n    youtubeId: dQw4w9WgXcQ\n    title: Video\n    channel: Channel\n    durationMinutes: 5\n    order: 0\n    objectiveIds: [objective]\nlessons: []\n")},
	})
	if err != nil {
		t.Fatalf("load video catalog: %v", err)
	}
	return catalog
}

func TestRemovedMigrationNeverDeletesHistoricalRows(t *testing.T) {
	db := openTestDatabase(t)
	seedHistory(t, db, identitySet{"lesson", "exercise", "test", "review"})
	migrations := parseMigrations(t, `version: 1
migrations:
  - entity: lesson
    from: ai-ml/module/lesson
    removed: true
`)
	if _, err := ApplyMigrations(context.Background(), db, catalog(t, identitySet{"lesson-new", "exercise", "test", "review"}), migrations); err != nil {
		t.Fatalf("apply removal ledger: %v", err)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM lesson_progress WHERE lesson_id = 'lesson'`, 1)
}

func TestApplyMigrationsRejectsReusedMigrationSourceBeforeWriting(t *testing.T) {
	db := openTestDatabase(t)
	seedHistory(t, db, identitySet{"lesson", "exercise", "test", "review"})
	migrations := parseMigrations(t, `version: 1
migrations:
  - entity: lesson
    from: ai-ml/module/lesson
    to: ai-ml/module/lesson-new
`)

	_, err := ApplyMigrations(context.Background(), db, catalog(t, identitySet{"lesson", "exercise", "test", "review"}), migrations)
	if err == nil || !strings.Contains(err.Error(), "migration sources are reserved") {
		t.Fatalf("expected reused migration source error, got %v", err)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM lesson_progress WHERE lesson_id = 'lesson' AND completed = 1`, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM lesson_progress WHERE lesson_id = 'lesson-new'`, 0)
}

func TestApplyMigrationsResolvesRenameChainsTransitively(t *testing.T) {
	db := openTestDatabase(t)
	seedHistory(t, db, identitySet{"lesson", "exercise", "test", "review"})
	current := catalog(t, identitySet{"lesson-new", "exercise", "test", "review"})
	migrations := parseMigrations(t, `version: 1
migrations:
  - entity: lesson
    from: ai-ml/module/lesson
    to: ai-ml/module/lesson-middle
  - entity: lesson
    from: ai-ml/module/lesson-middle
    to: ai-ml/module/lesson-new
`)
	if _, err := ApplyMigrations(context.Background(), db, current, migrations); err != nil {
		t.Fatalf("apply chained migrations: %v", err)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM lesson_progress WHERE lesson_id = 'lesson-new'`, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM lesson_progress WHERE lesson_id IN ('lesson', 'lesson-middle')`, 0)
}

func TestApplyMigrationsRejectsInvalidGraphsDeterministically(t *testing.T) {
	db := openTestDatabase(t)
	current := catalog(t, identitySet{"lesson-new", "exercise", "test", "review"})
	tests := []struct {
		name     string
		ledger   string
		contains string
	}{
		{
			name: "missing terminal",
			ledger: `version: 1
migrations:
  - entity: lesson
    from: ai-ml/module/lesson
    to: ai-ml/module/missing
`,
			contains: "missing terminal target lesson ai-ml/module/missing",
		},
		{
			name: "cycle",
			ledger: `version: 1
migrations:
  - entity: lesson
    from: ai-ml/module/lesson
    to: ai-ml/module/lesson-middle
  - entity: lesson
    from: ai-ml/module/lesson-middle
    to: ai-ml/module/lesson
`,
			contains: "identity migration cycle",
		},
		{
			name: "collision",
			ledger: `version: 1
migrations:
  - entity: lesson
    from: ai-ml/module/lesson
    to: ai-ml/module/lesson-new
  - entity: lesson
    from: ai-ml/module/another-lesson
    to: ai-ml/module/lesson-new
`,
			contains: "both target lesson ai-ml/module/lesson-new",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ApplyMigrations(context.Background(), db, current, parseMigrations(t, test.ledger))
			if err == nil || !strings.Contains(err.Error(), test.contains) {
				t.Fatalf("expected %q, got %v", test.contains, err)
			}
		})
	}
}

func TestResolveMigrationsComposesParentRenamesForDescendants(t *testing.T) {
	current := ownerCatalog(t, "course-c", "module-n")
	migrations := parseMigrations(t, `version: 1
migrations:
  - entity: course
    from: course-a
    to: course-b
  - entity: course
    from: course-b
    to: course-c
  - entity: module
    from: course-a/module-m
    to: course-b/module-n
`)
	resolved, err := resolveMigrations(current, migrations)
	if err != nil {
		t.Fatalf("resolve parent chain: %v", err)
	}
	for _, migration := range resolved {
		if migration.migration.Entity == curriculumidentity.ModuleKind && migration.destination.String() != "module course-c/module-n" {
			t.Fatalf("expected module terminal to include later course rename, got %s", migration.destination)
		}
	}
}

func TestApplyMigrationsRollsBackDatabaseIdentityCollision(t *testing.T) {
	db := openTestDatabase(t)
	seedHistory(t, db, identitySet{"lesson", "exercise", "test", "review"})
	if _, err := db.Exec(`INSERT INTO lesson_progress (user_id, course_id, module_id, lesson_id, completed, updated_at) VALUES ('00000000-0000-4000-8000-000000000001', 'ai-ml', 'module', 'lesson-new', 0, '2026-01-02')`); err != nil {
		t.Fatal(err)
	}
	current := catalog(t, identitySet{"lesson-new", "exercise", "test", "review"})
	migrations := parseMigrations(t, `version: 1
migrations:
  - entity: lesson
    from: ai-ml/module/lesson
    to: ai-ml/module/lesson-new
`)
	if _, err := ApplyMigrations(context.Background(), db, current, migrations); err == nil {
		t.Fatal("expected primary-key identity collision")
	}
	assertCount(t, db, `SELECT COUNT(*) FROM lesson_progress WHERE lesson_id = 'lesson' AND completed = 1`, 1)
}

func TestOpenReadOnlyCannotMutateDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "learner state.db")
	db, err := database.Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	readOnly, err := OpenReadOnly(context.Background(), path)
	if err != nil {
		t.Fatalf("open read-only: %v", err)
	}
	defer readOnly.Close()
	if _, err := readOnly.Exec(`INSERT INTO activities (kind, course_id, occurred_at) VALUES ('test', 'course', 'now')`); err == nil {
		t.Fatal("expected read-only database to reject writes")
	}
}

type identitySet struct {
	lesson   string
	exercise string
	test     string
	review   string
}

func catalog(t *testing.T, ids identitySet) *curriculum.Catalog {
	t.Helper()
	fsys := fstest.MapFS{
		"sources.yaml":                                         &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/ai-ml/course.yaml":                            &fstest.MapFile{Data: []byte("id: ai-ml\ntitle: Course\ndescription: Course\norder: 0\n")},
		"courses/ai-ml/modules/module/module.yaml":             &fstest.MapFile{Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective\n    title: Objective\n    description: Description\n    prerequisites: []\nvideos: []\nlessons:\n  - " + ids.lesson + "\n")},
		"courses/ai-ml/modules/module/lesson.mdx":              &fstest.MapFile{Data: []byte("---\nid: " + ids.lesson + "\ntitle: Lesson\nobjectiveIds:\n  - objective\nsourceIds: []\n---\n# Lesson\n")},
		"courses/ai-ml/modules/module/exercises/exercise.yaml": &fstest.MapFile{Data: []byte("id: " + ids.exercise + "\ntitle: Exercise\nlessonId: " + ids.lesson + "\norder: 0\nobjectiveIds:\n  - objective\nprompt: Prompt\nstarterCode: pass\ntests:\n  - id: " + ids.test + "\n    title: Test\n    visibility: visible\n    code: assert True\n")},
		"courses/ai-ml/modules/module/reviews/review.yaml":     &fstest.MapFile{Data: []byte("id: " + ids.review + "\norder: 0\nobjectiveIds:\n  - objective\nsourceLessonId: " + ids.lesson + "\nprompt: Prompt\nanswer: Answer\n")},
	}
	catalog, err := curriculum.Load(fsys)
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	return catalog
}

func ownerCatalog(t *testing.T, courseID, moduleID string) *curriculum.Catalog {
	t.Helper()
	fsys := fstest.MapFS{
		"sources.yaml":                           &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":             &fstest.MapFile{Data: []byte("id: " + courseID + "\ntitle: Course\ndescription: Course\norder: 0\n")},
		"courses/course/modules/mod/module.yaml": &fstest.MapFile{Data: []byte("id: " + moduleID + "\ntitle: Module\norder: 0\nobjectives: []\nvideos: []\nlessons: []\n")},
	}
	catalog, err := curriculum.Load(fsys)
	if err != nil {
		t.Fatalf("load owner catalog: %v", err)
	}
	return catalog
}

func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func seedHistory(t *testing.T, db *sql.DB, ids identitySet) {
	t.Helper()
	statements := []string{
		`INSERT INTO lesson_progress (user_id, course_id, module_id, lesson_id, completed, completed_at, updated_at) VALUES ('00000000-0000-4000-8000-000000000001', 'ai-ml', 'module', '` + ids.lesson + `', 1, '2026-01-01', '2026-01-01')`,
		`INSERT INTO activities (user_id, kind, course_id, module_id, lesson_id, occurred_at) VALUES ('00000000-0000-4000-8000-000000000001', 'lesson_completed', 'ai-ml', 'module', '` + ids.lesson + `', '2026-01-01')`,
		`INSERT INTO activities (user_id, kind, course_id, module_id, exercise_id, occurred_at) VALUES ('00000000-0000-4000-8000-000000000001', 'exercise_checked', 'ai-ml', 'module', '` + ids.exercise + `', '2026-01-02')`,
		`INSERT INTO activities (user_id, kind, course_id, module_id, review_item_id, occurred_at) VALUES ('00000000-0000-4000-8000-000000000001', 'review_completed', 'ai-ml', 'module', '` + ids.review + `', '2026-01-03')`,
		`INSERT INTO exercise_workspaces (user_id, course_id, module_id, exercise_id, code, updated_at) VALUES ('00000000-0000-4000-8000-000000000001', 'ai-ml', 'module', '` + ids.exercise + `', 'saved code', '2026-01-01')`,
		`INSERT INTO exercise_attempts (id, user_id, course_id, module_id, exercise_id, created_at, passed_count, failed_count, duration_ms, all_passed, code_snapshot) VALUES (1, '00000000-0000-4000-8000-000000000001', 'ai-ml', 'module', '` + ids.exercise + `', '2026-01-01', 1, 0, 5, 1, 'attempt code')`,
		`INSERT INTO exercise_test_results (user_id, attempt_id, test_id, status, message, duration_ms) VALUES ('00000000-0000-4000-8000-000000000001', 1, '` + ids.test + `', 'passed', 'history', 5)`,
		`INSERT INTO review_cards (user_id, course_id, module_id, review_item_id, due_at, stability, difficulty, scheduled_days, reps, lapses, state, last_review_at, remaining_steps, updated_at) VALUES ('00000000-0000-4000-8000-000000000001', 'ai-ml', 'module', '` + ids.review + `', '2026-01-02', 1, 5, 1, 2, 0, 2, '2026-01-01', 0, '2026-01-01')`,
		`INSERT INTO review_logs (id, user_id, course_id, module_id, review_item_id, reviewed_at, rating, previous_due, next_due, before_stability, after_stability, before_difficulty, after_difficulty, before_scheduled_days, after_scheduled_days, before_reps, after_reps, before_lapses, after_lapses, before_state, after_state, before_last_review_at, after_last_review_at, before_remaining_steps, after_remaining_steps) VALUES (1, '00000000-0000-4000-8000-000000000001', 'ai-ml', 'module', '` + ids.review + `', '2026-01-01', 'good', '2026-01-01', '2026-01-02', 1, 2, 5, 5, 0, 1, 1, 2, 0, 0, 1, 2, NULL, '2026-01-01', 1, 0)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("seed history: %v\n%s", err, statement)
		}
	}
}

func parseMigrations(t *testing.T, data string) []curriculumidentity.Migration {
	t.Helper()
	migrations, err := curriculumidentity.ParseMigrations([]byte(data))
	if err != nil {
		t.Fatalf("parse migrations: %v", err)
	}
	return migrations
}

func assertCount(t *testing.T, db *sql.DB, query string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(query).Scan(&got); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if got != want {
		t.Fatalf("expected %d rows, got %d for %s", want, got, query)
	}
}
