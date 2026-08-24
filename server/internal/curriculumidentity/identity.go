package curriculumidentity

import (
	"fmt"
	"sort"
	"strings"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

// Kind identifies an authored curriculum entity whose ID is persisted in
// learner state or deliberately protected ahead of imminent persistence. The
// qualified identity includes all owning entities.
type Kind string

const (
	CourseKind       Kind = "course"
	ModuleKind       Kind = "module"
	LessonKind       Kind = "lesson"
	ExerciseKind     Kind = "exercise"
	ExerciseTestKind Kind = "exercise-test"
	ReviewItemKind   Kind = "review-item"
	VideoKind        Kind = "video"
)

var kindArity = map[Kind]int{
	CourseKind:       1,
	ModuleKind:       2,
	LessonKind:       3,
	ExerciseKind:     3,
	ExerciseTestKind: 4,
	ReviewItemKind:   3,
	VideoKind:        3,
}

type Identity struct {
	Kind  Kind
	Parts []string
}

func (i Identity) String() string {
	return fmt.Sprintf("%s %s", i.Kind, strings.Join(i.Parts, "/"))
}

func (i Identity) key() string {
	return string(i.Kind) + "\x00" + strings.Join(i.Parts, "\x00")
}

// Snapshot derives persistence-sensitive identity from a validated catalog.
// It deliberately excludes authored entities with no persistence plan, such
// as worksheets and objectives. Videos are protected before learner video
// state begins storing their identity.
func Snapshot(catalog *curriculum.Catalog) []Identity {
	identities := make([]Identity, 0)
	for _, course := range catalog.Courses() {
		identities = append(identities, identity(CourseKind, course.ID))
		for _, module := range course.Modules {
			identities = append(identities, identity(ModuleKind, course.ID, module.ID))
			for _, video := range module.Videos {
				identities = append(identities, identity(VideoKind, course.ID, module.ID, video.ID))
			}
			for _, lesson := range module.Lessons {
				identities = append(identities, identity(LessonKind, course.ID, module.ID, lesson.ID))
			}
			for _, exercise := range module.Exercises {
				identities = append(identities, identity(ExerciseKind, course.ID, module.ID, exercise.ID))
				for _, test := range exercise.Tests {
					identities = append(identities, identity(ExerciseTestKind, course.ID, module.ID, exercise.ID, test.ID))
				}
			}
			for _, reviewItem := range module.ReviewItems {
				identities = append(identities, identity(ReviewItemKind, course.ID, module.ID, reviewItem.ID))
			}
		}
	}
	sortIdentities(identities)
	return identities
}

func identity(kind Kind, parts ...string) Identity {
	return Identity{Kind: kind, Parts: append([]string(nil), parts...)}
}

type BreakingChange struct {
	Identity Identity
	Reason   string
}

type Result struct {
	BreakingChanges   []BreakingChange
	Additions         []Identity
	AppliedMigrations []Migration
}

// Compare checks that every persistence-sensitive base identity still exists
// in head, after applying an explicitly authored rename or removal. New head
// identities are safe additions.
func Compare(base, head []Identity, migrations []Migration) Result {
	headByKey := make(map[string]Identity, len(head))
	for _, current := range head {
		headByKey[current.key()] = current
	}

	accountedHead := make(map[string]struct{}, len(base))
	claimedHead := make(map[string]string, len(base))
	applied := make(map[int]struct{})
	result := Result{}
	for _, previous := range base {
		if _, ok := headByKey[previous.key()]; ok {
			if owner, claimed := claimedHead[previous.key()]; claimed && owner != previous.key() {
				result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
					Identity: previous,
					Reason:   "identity is also the target of another base identity migration",
				})
				continue
			}
			accountedHead[previous.key()] = struct{}{}
			claimedHead[previous.key()] = previous.key()
			continue
		}
		migrationIndex := mostSpecificMigration(previous, migrations)
		if migrationIndex < 0 {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Identity: previous,
				Reason:   "missing without an identity migration",
			})
			continue
		}

		migration := migrations[migrationIndex]
		applied[migrationIndex] = struct{}{}
		if migration.Removed {
			continue
		}
		mapped := applyMigration(previous, migration)
		if _, ok := headByKey[mapped.key()]; !ok {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Identity: previous,
				Reason:   fmt.Sprintf("migration target %s does not exist", mapped),
			})
			continue
		}
		if owner, claimed := claimedHead[mapped.key()]; claimed && owner != previous.key() {
			result.BreakingChanges = append(result.BreakingChanges, BreakingChange{
				Identity: previous,
				Reason:   fmt.Sprintf("migration target %s is also claimed by another base identity", mapped),
			})
			continue
		}
		accountedHead[mapped.key()] = struct{}{}
		claimedHead[mapped.key()] = previous.key()
	}

	for _, current := range head {
		if _, ok := accountedHead[current.key()]; !ok {
			result.Additions = append(result.Additions, current)
		}
	}
	for index := range applied {
		result.AppliedMigrations = append(result.AppliedMigrations, migrations[index])
	}
	sort.Slice(result.BreakingChanges, func(i, j int) bool {
		return result.BreakingChanges[i].Identity.key() < result.BreakingChanges[j].Identity.key()
	})
	sortIdentities(result.Additions)
	sort.Slice(result.AppliedMigrations, func(i, j int) bool {
		return result.AppliedMigrations[i].key() < result.AppliedMigrations[j].key()
	})
	return result
}

func mostSpecificMigration(previous Identity, migrations []Migration) int {
	bestIndex, bestLength := -1, -1
	for index, migration := range migrations {
		if !migrationApplies(previous, migration) || len(migration.fromParts) <= bestLength {
			continue
		}
		bestIndex, bestLength = index, len(migration.fromParts)
	}
	return bestIndex
}

func migrationApplies(previous Identity, migration Migration) bool {
	if len(previous.Parts) < len(migration.fromParts) {
		return false
	}
	for index := range migration.fromParts {
		if previous.Parts[index] != migration.fromParts[index] {
			return false
		}
	}
	switch migration.Entity {
	case CourseKind:
		return true
	case ModuleKind:
		return previous.Kind != CourseKind
	case ExerciseKind:
		return previous.Kind == ExerciseKind || previous.Kind == ExerciseTestKind
	default:
		return previous.Kind == migration.Entity
	}
}

func applyMigration(previous Identity, migration Migration) Identity {
	parts := append([]string(nil), migration.toParts...)
	parts = append(parts, previous.Parts[len(migration.fromParts):]...)
	return Identity{Kind: previous.Kind, Parts: parts}
}

func sortIdentities(identities []Identity) {
	sort.Slice(identities, func(i, j int) bool { return identities[i].key() < identities[j].key() })
}
