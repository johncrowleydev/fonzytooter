package database

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/johncrowleydev/helix-academy/server/internal/auth"
)

func TestOpenCreatesParentAndMigratesFreshDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "helix-academy.db")

	db, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	var migrationCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM goose_db_version WHERE version_id BETWEEN 1 AND 12 AND is_applied = 1").Scan(&migrationCount); err != nil {
		t.Fatalf("query migration state: %v", err)
	}
	if migrationCount != 12 {
		t.Fatalf("expected all migrations, got %d rows", migrationCount)
	}
	for _, table := range []string{
		"lesson_progress",
		"video_progress",
		"activities",
		"exercise_workspaces",
		"exercise_attempts",
		"exercise_test_results",
		"review_cards",
		"review_logs",
		"tutor_conversations",
		"tutor_messages",
		"tutor_message_parts",
		"tutor_tool_calls",
		"tutor_conversation_memory",
		"users",
		"auth_sessions",
		"tutor_usage_windows",
	} {
		var tableCount int
		if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&tableCount); err != nil {
			t.Fatalf("query table %s: %v", table, err)
		}
		if tableCount != 1 {
			t.Fatalf("expected migrated table %s", table)
		}
	}
}

func TestOpenEnforcesForeignKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "helix-academy.db")
	db, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec("CREATE TABLE parents (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create parent table: %v", err)
	}
	if _, err := db.Exec("CREATE TABLE children (parent_id INTEGER REFERENCES parents(id))"); err != nil {
		t.Fatalf("create child table: %v", err)
	}
	if _, err := db.Exec("INSERT INTO children (parent_id) VALUES (99)"); err == nil {
		t.Fatal("expected foreign-key violation")
	}
}

func TestOpenConfiguresReplacementConnections(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "helix-academy.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	db.SetMaxIdleConns(0)
	for i := 0; i < 2; i++ {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("open replacement connection %d: %v", i+1, err)
		}

		var foreignKeys int
		if err := conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
			_ = conn.Close()
			t.Fatalf("query foreign-key setting on connection %d: %v", i+1, err)
		}
		if foreignKeys != 1 {
			_ = conn.Close()
			t.Fatalf("expected foreign keys on connection %d, got %d", i+1, foreignKeys)
		}

		var busyTimeout int
		if err := conn.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
			_ = conn.Close()
			t.Fatalf("query busy timeout on connection %d: %v", i+1, err)
		}
		if busyTimeout != busyTimeoutMilliseconds {
			_ = conn.Close()
			t.Fatalf("expected %d ms busy timeout on connection %d, got %d", busyTimeoutMilliseconds, i+1, busyTimeout)
		}

		if err := conn.Close(); err != nil {
			t.Fatalf("close replacement connection %d: %v", i+1, err)
		}
	}
}

func TestOpenMigratesIdempotently(t *testing.T) {
	path := filepath.Join(t.TempDir(), "helix-academy.db")

	first, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close first database: %v", err)
	}

	second, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })

	var migrationCount int
	if err := second.QueryRow("SELECT COUNT(*) FROM goose_db_version WHERE version_id BETWEEN 1 AND 12 AND is_applied = 1").Scan(&migrationCount); err != nil {
		t.Fatalf("query migration state: %v", err)
	}
	if migrationCount != 12 {
		t.Fatalf("expected all migrations after reopening, got %d", migrationCount)
	}
}

func TestUserScopedMigrationPreservesExistingLearnerState(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "helix-academy.db")
	legacyMigrations := fstest.MapFS{}
	entries, err := fs.ReadDir(embeddedMigrations, "migrations")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() >= "00010_" {
			continue
		}
		data, err := fs.ReadFile(embeddedMigrations, "migrations/"+entry.Name())
		if err != nil {
			t.Fatalf("read migration %s: %v", entry.Name(), err)
		}
		legacyMigrations[entry.Name()] = &fstest.MapFile{Data: data}
	}
	legacy, err := open(ctx, path, legacyMigrations)
	if err != nil {
		t.Fatalf("open pre-user-scope database: %v", err)
	}
	statements := []string{
		`INSERT INTO lesson_progress VALUES ('course', 'module', 'lesson', 1, '2026-08-20T00:00:00.000000000Z', '2026-08-20T00:00:00.000000000Z')`,
		`INSERT INTO activities (id, kind, course_id, module_id, lesson_id, occurred_at) VALUES (1, 'lesson_completed', 'course', 'module', 'lesson', '2026-08-20T00:00:00.000000000Z')`,
		`INSERT INTO exercise_workspaces VALUES ('course', 'module', 'exercise', 'saved code', '2026-08-20T00:00:00.000000000Z')`,
		`INSERT INTO exercise_attempts VALUES (1, 'course', 'module', 'exercise', '2026-08-20T00:00:00.000000000Z', 1, 0, 5, 1, 'saved code')`,
		`INSERT INTO exercise_test_results VALUES (1, 'visible', 'passed', '', 5)`,
		`INSERT INTO review_cards VALUES ('course', 'module', 'review', '2026-08-21T00:00:00.000000000Z', 1, 5, 1, 1, 0, 1, '2026-08-20T00:00:00.000000000Z', 0, '2026-08-20T00:00:00.000000000Z')`,
		`INSERT INTO review_logs (id, course_id, module_id, review_item_id, reviewed_at, rating, previous_due, next_due, before_stability, after_stability, before_difficulty, after_difficulty, before_scheduled_days, after_scheduled_days, before_reps, after_reps, before_lapses, after_lapses, before_state, after_state, before_last_review_at, after_last_review_at, before_remaining_steps, after_remaining_steps) VALUES (1, 'course', 'module', 'review', '2026-08-20T00:00:00.000000000Z', 'good', '2026-08-20T00:00:00.000000000Z', '2026-08-21T00:00:00.000000000Z', 0, 1, 0, 5, 0, 1, 0, 1, 0, 0, 0, 1, NULL, '2026-08-20T00:00:00.000000000Z', 0, 0)`,
		`INSERT INTO tutor_conversations VALUES ('conversation', 'course', 'Legacy conversation', '2026-08-20T00:00:00.000000000Z', '2026-08-20T00:00:00.000000000Z')`,
		`INSERT INTO tutor_messages (id, conversation_id, sequence, role, created_at) VALUES ('message', 'conversation', 1, 'assistant', '2026-08-20T00:00:00.000000000Z')`,
		`INSERT INTO tutor_message_parts VALUES ('message', 1, 'text', 'legacy message')`,
		`INSERT INTO tutor_tool_calls (id, conversation_id, message_id, sequence, request_id, name, arguments_json, status, created_at) VALUES ('call', 'conversation', 'message', 1, 'request', 'lookup', '{}', 'pending', '2026-08-20T00:00:00.000000000Z')`,
		`INSERT INTO tutor_conversation_memory VALUES ('conversation', 'legacy summary', '{}', 1, 1, '2026-08-20T00:00:00.000000000Z', '2026-08-20T00:00:00.000000000Z')`,
	}
	for _, statement := range statements {
		if _, err := legacy.ExecContext(ctx, statement); err != nil {
			_ = legacy.Close()
			t.Fatalf("seed pre-user-scope state: %v\n%s", err, statement)
		}
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close pre-user-scope database: %v", err)
	}

	upgraded, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("apply user-scoped migration: %v", err)
	}
	t.Cleanup(func() { _ = upgraded.Close() })
	for _, table := range []string{
		"lesson_progress", "activities", "exercise_workspaces", "exercise_attempts", "exercise_test_results",
		"review_cards", "review_logs", "tutor_conversations", "tutor_messages", "tutor_message_parts",
		"tutor_tool_calls", "tutor_conversation_memory",
	} {
		var count int
		if err := upgraded.QueryRow("SELECT COUNT(*) FROM "+table+" WHERE user_id = ?", auth.BootstrapUserID).Scan(&count); err != nil {
			t.Fatalf("read migrated %s ownership: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("migrated %s rows for bootstrap user = %d, want 1", table, count)
		}
	}
	rows, err := upgraded.Query("PRAGMA foreign_key_check")
	if err != nil {
		t.Fatalf("check migrated foreign keys: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatal("user-scoped migration left a foreign-key violation")
	}
}

func TestOpenNormalizesLegacyHistoryTimestamps(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "helix-academy.db")
	legacyMigrations := fstest.MapFS{}
	entries, err := fs.ReadDir(embeddedMigrations, "migrations")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() >= "00008_" {
			continue
		}
		data, err := fs.ReadFile(embeddedMigrations, "migrations/"+entry.Name())
		if err != nil {
			t.Fatalf("read migration %s: %v", entry.Name(), err)
		}
		legacyMigrations[entry.Name()] = &fstest.MapFile{Data: data}
	}
	legacy, err := open(ctx, path, legacyMigrations)
	if err != nil {
		t.Fatalf("open pre-normalization database: %v", err)
	}
	statements := []string{
		`INSERT INTO activities (kind, course_id, occurred_at) VALUES ('legacy', 'course', '2026-08-20T12:30:00.1Z')`,
		`INSERT INTO exercise_attempts (course_id, module_id, exercise_id, created_at, passed_count, failed_count, duration_ms, all_passed, code_snapshot) VALUES ('course', 'module', 'exercise', '2026-08-20T12:30:00.11Z', 1, 0, 0, 1, '')`,
		`INSERT INTO review_cards (course_id, module_id, review_item_id, due_at, stability, difficulty, scheduled_days, reps, lapses, state, remaining_steps, updated_at) VALUES ('course', 'module', 'review', '2026-08-20T12:30:00Z', 0, 0, 0, 0, 0, 0, 0, '2026-08-20T12:30:00Z')`,
		`INSERT INTO review_logs (course_id, module_id, review_item_id, reviewed_at, rating, previous_due, next_due, before_stability, after_stability, before_difficulty, after_difficulty, before_scheduled_days, after_scheduled_days, before_reps, after_reps, before_lapses, after_lapses, before_state, after_state, before_remaining_steps, after_remaining_steps) VALUES ('course', 'module', 'review', '2026-08-20T12:30:00Z', 'good', '2026-08-20T12:30:00Z', '2026-08-21T12:30:00Z', 0, 1, 0, 1, 0, 1, 0, 1, 0, 0, 0, 1, 0, 0)`,
	}
	for _, statement := range statements {
		if _, err := legacy.ExecContext(ctx, statement); err != nil {
			_ = legacy.Close()
			t.Fatalf("seed legacy timestamp: %v", err)
		}
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close pre-normalization database: %v", err)
	}

	upgraded, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("upgrade database: %v", err)
	}
	t.Cleanup(func() { _ = upgraded.Close() })
	for table, test := range map[string]struct {
		column string
		want   string
	}{
		"activities":        {"occurred_at", "2026-08-20T12:30:00.100000000Z"},
		"exercise_attempts": {"created_at", "2026-08-20T12:30:00.110000000Z"},
		"review_logs":       {"reviewed_at", "2026-08-20T12:30:00.000000000Z"},
	} {
		var got string
		if err := upgraded.QueryRow("SELECT " + test.column + " FROM " + table + " LIMIT 1").Scan(&got); err != nil {
			t.Fatalf("read normalized %s: %v", table, err)
		}
		if got != test.want {
			t.Fatalf("normalized %s timestamp = %q, want %q", table, got, test.want)
		}
	}
}

func TestOpenRejectsInvalidPath(t *testing.T) {
	parentFile := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parentFile, []byte("file"), 0o600); err != nil {
		t.Fatalf("create parent file: %v", err)
	}

	_, err := Open(context.Background(), filepath.Join(parentFile, "helix-academy.db"))
	if err == nil || !strings.Contains(err.Error(), "create database directory") {
		t.Fatalf("expected directory creation error, got %v", err)
	}
}

func TestOpenFailsForInvalidMigration(t *testing.T) {
	migrations := fstest.MapFS{
		"00001_invalid.sql": {Data: []byte("-- +goose Up\nTHIS IS NOT SQL;")},
	}

	_, err := open(context.Background(), filepath.Join(t.TempDir(), "helix-academy.db"), migrations)
	if err == nil || !strings.Contains(err.Error(), "apply SQLite migrations") {
		t.Fatalf("expected migration error, got %v", err)
	}
}

func TestClosedDatabaseRejectsPing(t *testing.T) {
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "helix-academy.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	if err := db.Ping(); err == nil {
		t.Fatal("expected ping to fail after close")
	}
}
