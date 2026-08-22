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
	"github.com/johncrowleydev/fonzytooter/server/internal/tutor"
	openrouterprovider "github.com/johncrowleydev/fonzytooter/server/internal/tutor/openrouter"
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

func TestConfiguredTutorProviderPreservesDisabledPath(t *testing.T) {
	provider, err := configuredTutorProvider(config.Config{})
	if err != nil {
		t.Fatalf("configure disabled tutor: %v", err)
	}
	if _, ok := provider.(*tutor.UnavailableProvider); !ok {
		t.Fatalf("expected unavailable provider, got %T", provider)
	}
}

func TestConfiguredTutorProviderBuildsOpenRouterOnlyWhenComplete(t *testing.T) {
	cfg := config.Config{OpenRouter: config.OpenRouterConfig{
		APIKey: "test-key", Model: "vendor/exact-model", BaseURL: "https://openrouter.test/v1",
	}}
	provider, err := configuredTutorProvider(cfg)
	if err != nil {
		t.Fatalf("configure OpenRouter: %v", err)
	}
	if _, ok := provider.(*openrouterprovider.Provider); !ok {
		t.Fatalf("expected OpenRouter provider, got %T", provider)
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
