package curriculumidentity

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

func TestSnapshotContainsPersistenceSensitiveAuthoredIDs(t *testing.T) {
	got := Snapshot(mustCatalog(t, moduleFiles("lesson", "exercise", "review", "test")))
	identities := make([]string, len(got))
	for index, current := range got {
		identities[index] = current.String()
	}
	want := []string{
		"course ai-ml",
		"exercise ai-ml/module/exercise",
		"exercise-test ai-ml/module/exercise/test",
		"lesson ai-ml/module/lesson",
		"module ai-ml/module",
		"review-item ai-ml/module/review",
	}
	if strings.Join(identities, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected identity policy:\n%s", strings.Join(identities, "\n"))
	}
}

func TestCompareAllowsAdditionsAndDetectsDeterministicRemovals(t *testing.T) {
	base := snapshot(t, moduleFiles("lesson", "exercise", "review", "test"))
	head := snapshot(t, moduleFiles("lesson-new", "exercise-new", "review-new", "test-new"))

	result := Compare(base, head, nil)
	if len(result.Additions) != 4 {
		t.Fatalf("expected four additions, got %#v", result.Additions)
	}
	got := make([]string, len(result.BreakingChanges))
	for index, change := range result.BreakingChanges {
		got[index] = change.Identity.String()
	}
	want := []string{
		"exercise ai-ml/module/exercise",
		"exercise-test ai-ml/module/exercise/test",
		"lesson ai-ml/module/lesson",
		"review-item ai-ml/module/review",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected deterministic changes:\n%s", strings.Join(got, "\n"))
	}
}

func TestCompareAppliesExplicitRenamesAndRemovals(t *testing.T) {
	base := snapshot(t, moduleFiles("lesson", "exercise", "review", "test"))
	head := snapshot(t, moduleFiles("lesson-new", "exercise-new", "review-new", "test"))
	migrations := parse(t, `version: 1
migrations:
  - entity: lesson
    from: ai-ml/module/lesson
    to: ai-ml/module/lesson-new
  - entity: exercise
    from: ai-ml/module/exercise
    to: ai-ml/module/exercise-new
  - entity: review-item
    from: ai-ml/module/review
    removed: true
`)

	result := Compare(base, head, migrations)
	if len(result.BreakingChanges) != 0 {
		t.Fatalf("expected accounted changes, got %#v", result.BreakingChanges)
	}
	if len(result.AppliedMigrations) != 3 {
		t.Fatalf("expected three applied migrations, got %#v", result.AppliedMigrations)
	}
	// The exercise migration also accounts for the persisted test identity.
	if len(result.Additions) != 1 || result.Additions[0].String() != "review-item ai-ml/module/review-new" {
		t.Fatalf("expected only replacement for an explicitly removed item to be an addition, got %#v", result.Additions)
	}
}

func TestCompareParentRenameAccountsForDescendants(t *testing.T) {
	base := snapshot(t, courseFiles("ai-ml", "module", "lesson"))
	head := snapshot(t, courseFiles("machine-learning", "foundations", "lesson"))
	migrations := parse(t, `version: 1
migrations:
  - entity: course
    from: ai-ml
    to: machine-learning
  - entity: module
    from: ai-ml/module
    to: machine-learning/foundations
`)

	result := Compare(base, head, migrations)
	if len(result.BreakingChanges) != 0 || len(result.Additions) != 0 {
		t.Fatalf("expected parent mappings to preserve descendants, got %#v", result)
	}
}

func TestCompareRejectsMissingMigrationTarget(t *testing.T) {
	base := snapshot(t, moduleFiles("lesson", "exercise", "review", "test"))
	head := snapshot(t, moduleFiles("other", "exercise", "review", "test"))
	migrations := parse(t, `version: 1
migrations:
  - entity: lesson
    from: ai-ml/module/lesson
    to: ai-ml/module/not-authored
`)

	result := Compare(base, head, migrations)
	if len(result.BreakingChanges) != 1 || !strings.Contains(result.BreakingChanges[0].Reason, "does not exist") {
		t.Fatalf("expected missing target error, got %#v", result.BreakingChanges)
	}
}

func TestCompareRejectsMigrationThatCollapsesTwoBaseIdentities(t *testing.T) {
	base := []Identity{identity(LessonKind, "course", "module", "first"), identity(LessonKind, "course", "module", "second")}
	head := []Identity{identity(LessonKind, "course", "module", "second")}
	migrations := parse(t, `version: 1
migrations:
  - entity: lesson
    from: course/module/first
    to: course/module/second
`)

	result := Compare(base, head, migrations)
	if len(result.BreakingChanges) != 1 || !strings.Contains(result.BreakingChanges[0].Reason, "another base identity") {
		t.Fatalf("expected identity collision, got %#v", result.BreakingChanges)
	}
}

func TestParseMigrationsRejectsAmbiguousAndMalformedEntries(t *testing.T) {
	for name, input := range map[string]string{
		"missing disposition": "version: 1\nmigrations:\n  - entity: lesson\n    from: ai-ml/module/lesson\n",
		"both dispositions":   "version: 1\nmigrations:\n  - entity: lesson\n    from: ai-ml/module/lesson\n    to: ai-ml/module/new\n    removed: true\n",
		"wrong qualification": "version: 1\nmigrations:\n  - entity: lesson\n    from: lesson\n    removed: true\n",
		"unknown field":       "version: 1\nmigrations: []\nextra: true\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseMigrations([]byte(input)); err == nil {
				t.Fatal("expected parse error")
			}
		})
	}
}

func TestValidateAppendOnlyAllowsAdditions(t *testing.T) {
	base := parse(t, "version: 1\nmigrations:\n  - entity: lesson\n    from: course/module/old\n    to: course/module/new\n")
	head := parse(t, "version: 1\nmigrations:\n  - entity: lesson\n    from: course/module/old\n    to: course/module/new\n  - entity: exercise\n    from: course/module/old\n    removed: true\n")
	if err := ValidateAppendOnly(base, head); err != nil {
		t.Fatalf("validate append-only ledger: %v", err)
	}
}

func TestValidateAppendOnlyRejectsRemovedAndRewrittenHistoryDeterministically(t *testing.T) {
	base := parse(t, "version: 1\nmigrations:\n  - entity: lesson\n    from: course/module/removed\n    removed: true\n  - entity: lesson\n    from: course/module/rewritten\n    to: course/module/new\n")
	head := parse(t, "version: 1\nmigrations:\n  - entity: lesson\n    from: course/module/rewritten\n    removed: true\n")
	err := ValidateAppendOnly(base, head)
	want := "identity migration ledger is append-only: removed entries: lesson course/module/removed; rewritten entries: lesson course/module/rewritten"
	if err == nil || err.Error() != want {
		t.Fatalf("append-only error = %v, want %q", err, want)
	}
}

func TestValidateNewMigrationsAppliedRejectsPreauthorizationDeterministically(t *testing.T) {
	head := parse(t, `version: 1
migrations:
  - entity: lesson
    from: course/module/zeta
    removed: true
  - entity: lesson
    from: course/module/alpha
    to: course/module/beta
`)
	err := ValidateNewMigrationsApplied(nil, head, nil)
	want := "new identity migrations must be exercised by the current base-to-head change: lesson course/module/alpha, lesson course/module/zeta"
	if err == nil || err.Error() != want {
		t.Fatalf("migration usage error = %v, want %q", err, want)
	}
}

func TestValidateNewMigrationsAppliedAllowsCurrentBreakingChange(t *testing.T) {
	base := parse(t, "version: 1\nmigrations: []\n")
	head := parse(t, "version: 1\nmigrations:\n  - entity: lesson\n    from: course/module/old\n    to: course/module/new\n")
	if err := ValidateNewMigrationsApplied(base, head, head); err != nil {
		t.Fatalf("validate exercised migration: %v", err)
	}
}

func TestValidateReservedMigrationSourcesRejectsReusedIdentity(t *testing.T) {
	migrations := parse(t, "version: 1\nmigrations:\n  - entity: lesson\n    from: course/module/old\n    to: course/module/new\n")
	head := []Identity{identity(LessonKind, "course", "module", "old")}
	err := ValidateReservedMigrationSources(head, migrations)
	want := "identity migration sources are reserved and cannot exist in the current curriculum: lesson course/module/old"
	if err == nil || err.Error() != want {
		t.Fatalf("reserved source error = %v, want %q", err, want)
	}
}

func parse(t *testing.T, input string) []Migration {
	t.Helper()
	migrations, err := ParseMigrations([]byte(input))
	if err != nil {
		t.Fatalf("parse migrations: %v", err)
	}
	return migrations
}

func snapshot(t *testing.T, files fstest.MapFS) []Identity {
	t.Helper()
	return Snapshot(mustCatalog(t, files))
}

func mustCatalog(t *testing.T, files fstest.MapFS) *curriculum.Catalog {
	t.Helper()
	catalog, err := curriculum.Load(files)
	if err != nil {
		t.Fatalf("load curriculum: %v", err)
	}
	return catalog
}

func moduleFiles(lessonID, exerciseID, reviewID, testID string) fstest.MapFS {
	return fstest.MapFS{
		"sources.yaml":                                         &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/ai-ml/course.yaml":                            &fstest.MapFile{Data: []byte("id: ai-ml\ntitle: AI ML\ndescription: Course\norder: 0\n")},
		"courses/ai-ml/modules/module/module.yaml":             &fstest.MapFile{Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective\n    title: Objective\n    description: Objective description\n    prerequisites: []\nvideos: []\nlessons:\n  - " + lessonID + "\n")},
		"courses/ai-ml/modules/module/lesson.mdx":              &fstest.MapFile{Data: []byte("---\nid: " + lessonID + "\ntitle: Lesson\nobjectiveIds:\n  - objective\nsourceIds: []\n---\n# Lesson\n")},
		"courses/ai-ml/modules/module/exercises/exercise.yaml": &fstest.MapFile{Data: []byte("id: " + exerciseID + "\ntitle: Exercise\nlessonId: " + lessonID + "\norder: 0\nobjectiveIds:\n  - objective\nprompt: Prompt\nstarterCode: pass\ntests:\n  - id: " + testID + "\n    title: Test\n    visibility: visible\n    code: assert True\n")},
		"courses/ai-ml/modules/module/reviews/review.yaml":     &fstest.MapFile{Data: []byte("id: " + reviewID + "\norder: 0\nobjectiveIds:\n  - objective\nsourceLessonId: " + lessonID + "\nprompt: Prompt\nanswer: Answer\n")},
	}
}

func courseFiles(courseID, moduleID, lessonID string) fstest.MapFS {
	return fstest.MapFS{
		"sources.yaml":                              &fstest.MapFile{Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                &fstest.MapFile{Data: []byte("id: " + courseID + "\ntitle: Course\ndescription: Course\norder: 0\n")},
		"courses/course/modules/module/module.yaml": &fstest.MapFile{Data: []byte("id: " + moduleID + "\ntitle: Module\norder: 0\nobjectives: []\nvideos: []\nlessons:\n  - " + lessonID + "\n")},
		"courses/course/modules/module/lesson.mdx":  &fstest.MapFile{Data: []byte("---\nid: " + lessonID + "\ntitle: Lesson\nobjectiveIds: []\nsourceIds: []\n---\n# Lesson\n")},
	}
}
