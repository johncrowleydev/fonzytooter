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
	if err := db.QueryRow("SELECT COUNT(*) FROM goose_db_version WHERE version_id = 1 AND is_applied = 1").Scan(&migrationCount); err != nil {
		t.Fatalf("query migration state: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected one applied migration, got %d", migrationCount)
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
	if err := second.QueryRow("SELECT COUNT(*) FROM goose_db_version WHERE version_id = 1 AND is_applied = 1").Scan(&migrationCount); err != nil {
		t.Fatalf("query migration state: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected one applied migration after reopening, got %d", migrationCount)
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
