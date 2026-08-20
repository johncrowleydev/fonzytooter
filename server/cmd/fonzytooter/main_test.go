package main

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johncrowleydev/fonzytooter/server/internal/config"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
)

func TestPrepareServerFailsWhenDatabaseMigrationFails(t *testing.T) {
	cfg := config.Config{
		Address:        ":0",
		DatabasePath:   "learner.db",
		CurriculumPath: "unused-because-database-opens-first",
	}
	migrationError := errors.New("apply SQLite migrations: invalid migration")

	server, db, err := prepareServer(context.Background(), cfg, func(context.Context, string) (*sql.DB, error) {
		return nil, migrationError
	})

	if server != nil || db != nil {
		t.Fatalf("expected no server or database after migration failure, got %#v, %#v", server, db)
	}
	if !errors.Is(err, migrationError) || !strings.Contains(err.Error(), "prepare learner database") {
		t.Fatalf("expected contextual migration startup error, got %v", err)
	}
}

func TestRunClosesDatabaseOnShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.Config{
		Address:        "127.0.0.1:0",
		DatabasePath:   filepath.Join(t.TempDir(), "fonzytooter.db"),
		CurriculumPath: filepath.Join("..", "..", "..", "curriculum"),
	}
	var opened *sql.DB

	err := run(ctx, cfg, func(ctx context.Context, path string) (*sql.DB, error) {
		var err error
		opened, err = database.Open(ctx, path)
		cancel()
		return opened, err
	})

	if err != nil {
		t.Fatalf("run server: %v", err)
	}
	if opened == nil {
		t.Fatal("expected database to open")
	}
	if err := opened.Ping(); err == nil {
		t.Fatal("expected database to be closed after shutdown")
	}
}
