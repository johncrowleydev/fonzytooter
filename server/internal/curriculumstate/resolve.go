package curriculumstate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/curriculumidentity"
)

type resolvedMigration struct {
	migration   curriculumidentity.Migration
	destination curriculumidentity.Identity
	removed     bool
	path        []string
}

func resolveMigrations(catalog *curriculum.Catalog, migrations []curriculumidentity.Migration) ([]resolvedMigration, error) {
	predecessors := make(map[string][]string)
	for _, migration := range migrations {
		sourceKey := migration.Source().String()
		if destination, ok := migration.Destination(); ok {
			destinationKey := destination.String()
			predecessors[destinationKey] = append(predecessors[destinationKey], sourceKey)
		}
	}
	var destinations []string
	for destination := range predecessors {
		destinations = append(destinations, destination)
	}
	sort.Strings(destinations)
	for _, destination := range destinations {
		sources := predecessors[destination]
		sort.Strings(sources)
		if len(sources) > 1 {
			return nil, fmt.Errorf("identity migration collision: %s and %s both target %s", sources[0], sources[1], destination)
		}
	}

	current := make(map[string]struct{})
	for _, identity := range curriculumidentity.Snapshot(catalog) {
		current[identity.String()] = struct{}{}
	}
	ordered := append([]curriculumidentity.Migration(nil), migrations...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Source().String() < ordered[j].Source().String() })
	resolved := make([]resolvedMigration, 0, len(ordered))
	for _, migration := range ordered {
		if migration.Removed {
			resolved = append(resolved, resolvedMigration{migration: migration, removed: true, path: []string{migration.Source().String()}})
			continue
		}
		visited := map[string]struct{}{migration.Source().String(): {}}
		chain := []string{migration.Source().String()}
		destination := migration.Source()
		for {
			next, continues := mostSpecificApplicable(destination, migrations)
			if !continues {
				key := destination.String()
				if _, exists := current[key]; !exists {
					return nil, fmt.Errorf("identity migration %s has missing terminal target %s", migration.Source(), destination)
				}
				resolved = append(resolved, resolvedMigration{migration: migration, destination: destination, path: chain})
				break
			}
			if next.Removed {
				resolved = append(resolved, resolvedMigration{migration: migration, removed: true, path: chain})
				break
			}
			destination = applyIdentityMigration(destination, next)
			key := destination.String()
			if _, cycle := visited[key]; cycle {
				chain = append(chain, key)
				return nil, fmt.Errorf("identity migration cycle: %s", strings.Join(chain, " -> "))
			}
			visited[key] = struct{}{}
			chain = append(chain, key)
		}
	}
	if err := validateResolvedCollisions(resolved); err != nil {
		return nil, err
	}
	return resolved, nil
}

func mostSpecificApplicable(identity curriculumidentity.Identity, migrations []curriculumidentity.Migration) (curriculumidentity.Migration, bool) {
	bestLength := -1
	var best curriculumidentity.Migration
	for _, migration := range migrations {
		source := migration.Source()
		if len(identity.Parts) < len(source.Parts) || len(source.Parts) <= bestLength {
			continue
		}
		matches := true
		for index := range source.Parts {
			if identity.Parts[index] != source.Parts[index] {
				matches = false
				break
			}
		}
		if !matches || !kindApplies(identity.Kind, migration.Entity) {
			continue
		}
		best, bestLength = migration, len(source.Parts)
	}
	return best, bestLength >= 0
}

func kindApplies(identity, migration curriculumidentity.Kind) bool {
	switch migration {
	case curriculumidentity.CourseKind:
		return true
	case curriculumidentity.ModuleKind:
		return identity != curriculumidentity.CourseKind
	case curriculumidentity.ExerciseKind:
		return identity == curriculumidentity.ExerciseKind || identity == curriculumidentity.ExerciseTestKind
	default:
		return identity == migration
	}
}

func applyIdentityMigration(identity curriculumidentity.Identity, migration curriculumidentity.Migration) curriculumidentity.Identity {
	source := migration.Source()
	destination, _ := migration.Destination()
	parts := append([]string(nil), destination.Parts...)
	parts = append(parts, identity.Parts[len(source.Parts):]...)
	return curriculumidentity.Identity{Kind: identity.Kind, Parts: parts}
}

func validateResolvedCollisions(resolved []resolvedMigration) error {
	byTerminal := make(map[string][]resolvedMigration)
	for _, migration := range resolved {
		if !migration.removed {
			byTerminal[migration.destination.String()] = append(byTerminal[migration.destination.String()], migration)
		}
	}
	var terminals []string
	for terminal := range byTerminal {
		terminals = append(terminals, terminal)
	}
	sort.Strings(terminals)
	for _, terminal := range terminals {
		candidates := byTerminal[terminal]
		for left := 0; left < len(candidates); left++ {
			for right := left + 1; right < len(candidates); right++ {
				leftSource := candidates[left].migration.Source().String()
				rightSource := candidates[right].migration.Source().String()
				if contains(candidates[left].path, rightSource) || contains(candidates[right].path, leftSource) {
					continue
				}
				if rightSource < leftSource {
					leftSource, rightSource = rightSource, leftSource
				}
				return fmt.Errorf("identity migration collision: %s and %s both resolve to %s", leftSource, rightSource, terminal)
			}
		}
	}
	return nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
