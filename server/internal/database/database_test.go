package database

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestOpenCreatesParentAndMigratesFreshDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "fonzytooter.db")

	db, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	var migrationCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM goose_db_version WHERE version_id BETWEEN 1 AND 7 AND is_applied = 1").Scan(&migrationCount); err != nil {
		t.Fatalf("query migration state: %v", err)
	}
	if migrationCount != 7 {
		t.Fatalf("expected all migrations, got %d rows", migrationCount)
	}
	for _, table := range []string{
		"lesson_progress",
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
	path := filepath.Join(t.TempDir(), "fonzytooter.db")
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
	db, err := Open(ctx, filepath.Join(t.TempDir(), "fonzytooter.db"))
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
	path := filepath.Join(t.TempDir(), "fonzytooter.db")

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
	if err := second.QueryRow("SELECT COUNT(*) FROM goose_db_version WHERE version_id BETWEEN 1 AND 7 AND is_applied = 1").Scan(&migrationCount); err != nil {
		t.Fatalf("query migration state: %v", err)
	}
	if migrationCount != 7 {
		t.Fatalf("expected all migrations after reopening, got %d", migrationCount)
	}
}

func TestOpenRejectsInvalidPath(t *testing.T) {
	parentFile := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parentFile, []byte("file"), 0o600); err != nil {
		t.Fatalf("create parent file: %v", err)
	}

	_, err := Open(context.Background(), filepath.Join(parentFile, "fonzytooter.db"))
	if err == nil || !strings.Contains(err.Error(), "create database directory") {
		t.Fatalf("expected directory creation error, got %v", err)
	}
}

func TestOpenFailsForInvalidMigration(t *testing.T) {
	migrations := fstest.MapFS{
		"00001_invalid.sql": {Data: []byte("-- +goose Up\nTHIS IS NOT SQL;")},
	}

	_, err := open(context.Background(), filepath.Join(t.TempDir(), "fonzytooter.db"), migrations)
	if err == nil || !strings.Contains(err.Error(), "apply SQLite migrations") {
		t.Fatalf("expected migration error, got %v", err)
	}
}

func TestClosedDatabaseRejectsPing(t *testing.T) {
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "fonzytooter.db"))
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
