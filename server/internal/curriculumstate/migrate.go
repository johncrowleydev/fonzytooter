package curriculumstate

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/johncrowleydev/helix-academy/server/internal/curriculum"
	"github.com/johncrowleydev/helix-academy/server/internal/curriculumidentity"
)

type MigrationUpdate struct {
	Entity  curriculumidentity.Kind
	From    string
	To      string
	Updates int64
}

// ApplyMigrations updates persisted identity columns in one transaction. It
// never deletes state: entries marked removed are deliberately ignored so the
// audit continues to report the preserved historical rows.
func ApplyMigrations(ctx context.Context, db *sql.DB, catalog *curriculum.Catalog, migrations []curriculumidentity.Migration) ([]MigrationUpdate, error) {
	resolved, err := resolveMigrations(catalog, migrations)
	if err != nil {
		return nil, err
	}
	renames := make([]resolvedMigration, 0, len(resolved))
	for _, migration := range resolved {
		if !migration.removed {
			renames = append(renames, migration)
		}
	}
	sort.Slice(renames, func(i, j int) bool {
		left, right := renames[i].migration.Source(), renames[j].migration.Source()
		if len(left.Parts) != len(right.Parts) {
			return len(left.Parts) > len(right.Parts)
		}
		return left.String() < right.String()
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin curriculum identity migration: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `PRAGMA defer_foreign_keys = ON`); err != nil {
		return nil, fmt.Errorf("defer curriculum identity foreign keys: %w", err)
	}

	totals := make(map[string]*MigrationUpdate, len(renames))
	for _, resolvedMigration := range renames {
		migration := resolvedMigration.migration
		totals[string(migration.Entity)+"\x00"+migration.From] = &MigrationUpdate{
			Entity: migration.Entity,
			From:   migration.From,
			To:     joinParts(resolvedMigration.destination.Parts),
		}
	}
	for _, resolvedMigration := range renames {
		migration := resolvedMigration.migration
		updates, err := applyMigration(ctx, tx, migration, resolvedMigration.destination)
		if err != nil {
			return nil, fmt.Errorf("migrate %s %s to %s: %w", migration.Entity, migration.From, resolvedMigration.destination, err)
		}
		totals[string(migration.Entity)+"\x00"+migration.From].Updates += updates
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit curriculum identity migration: %w", err)
	}
	updates := make([]MigrationUpdate, 0, len(totals))
	for _, update := range totals {
		updates = append(updates, *update)
	}
	sort.Slice(updates, func(i, j int) bool {
		if updates[i].Entity != updates[j].Entity {
			return updates[i].Entity < updates[j].Entity
		}
		return updates[i].From < updates[j].From
	})
	return updates, nil
}

func applyMigration(ctx context.Context, tx *sql.Tx, migration curriculumidentity.Migration, destination curriculumidentity.Identity) (int64, error) {
	from := migration.Source().Parts
	to := destination.Parts
	var statements []statement
	switch migration.Entity {
	case curriculumidentity.CourseKind:
		for _, table := range identityTables {
			statements = append(statements, statement{`UPDATE ` + table + ` SET course_id = ? WHERE course_id = ?`, []any{to[0], from[0]}})
		}
	case curriculumidentity.ModuleKind:
		for _, table := range identityTables {
			statements = append(statements, statement{`UPDATE ` + table + ` SET course_id = ?, module_id = ? WHERE course_id = ? AND module_id = ?`, []any{to[0], to[1], from[0], from[1]}})
		}
	case curriculumidentity.LessonKind:
		for _, table := range []string{"lesson_progress", "activities"} {
			statements = append(statements, statement{`UPDATE ` + table + ` SET course_id = ?, module_id = ?, lesson_id = ? WHERE course_id = ? AND module_id = ? AND lesson_id = ?`, []any{to[0], to[1], to[2], from[0], from[1], from[2]}})
		}
	case curriculumidentity.VideoKind:
		for _, table := range []string{"video_progress", "activities"} {
			statements = append(statements, statement{`UPDATE ` + table + ` SET course_id = ?, module_id = ?, video_id = ? WHERE course_id = ? AND module_id = ? AND video_id = ?`, []any{to[0], to[1], to[2], from[0], from[1], from[2]}})
		}
	case curriculumidentity.ExerciseKind:
		for _, table := range []string{"exercise_workspaces", "exercise_attempts", "activities"} {
			statements = append(statements, statement{`UPDATE ` + table + ` SET course_id = ?, module_id = ?, exercise_id = ? WHERE course_id = ? AND module_id = ? AND exercise_id = ?`, []any{to[0], to[1], to[2], from[0], from[1], from[2]}})
		}
	case curriculumidentity.ExerciseTestKind:
		statements = append(statements, statement{`
			UPDATE exercise_test_results SET test_id = ?
			WHERE test_id = ? AND attempt_id IN (
				SELECT id FROM exercise_attempts
				WHERE course_id = ? AND module_id = ? AND exercise_id = ?
			)`, []any{to[3], from[3], from[0], from[1], from[2]}})
	case curriculumidentity.ReviewItemKind:
		for _, table := range []string{"review_cards", "review_logs", "activities"} {
			statements = append(statements, statement{`UPDATE ` + table + ` SET course_id = ?, module_id = ?, review_item_id = ? WHERE course_id = ? AND module_id = ? AND review_item_id = ?`, []any{to[0], to[1], to[2], from[0], from[1], from[2]}})
		}
	default:
		return 0, fmt.Errorf("unsupported entity %q", migration.Entity)
	}

	var updates int64
	for _, query := range statements {
		result, err := tx.ExecContext(ctx, query.sql, query.args...)
		if err != nil {
			return 0, err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		updates += count
	}
	return updates, nil
}

var identityTables = []string{
	"activities",
	"exercise_attempts",
	"exercise_workspaces",
	"lesson_progress",
	"video_progress",
	"review_cards",
	"review_logs",
}

type statement struct {
	sql  string
	args []any
}

func joinParts(parts []string) string {
	result := ""
	for index, part := range parts {
		if index != 0 {
			result += "/"
		}
		result += part
	}
	return result
}
