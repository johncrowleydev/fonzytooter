package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

const (
	busyTimeoutMilliseconds = 5_000
	maxOpenConnections      = 1
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

// Open prepares a SQLite database and applies all embedded migrations.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	migrations, err := fs.Sub(embeddedMigrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("open embedded migrations: %w", err)
	}

	return open(ctx, path, migrations)
}

func open(ctx context.Context, path string, migrations fs.FS) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("database path is empty")
	}

	parent := filepath.Dir(path)
	if parent != "." {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return nil, fmt.Errorf("create database directory %q: %w", parent, err)
		}
	}

	dsn := fmt.Sprintf("%s?_foreign_keys=on&_busy_timeout=%d", path, busyTimeoutMilliseconds)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database %q: %w", path, err)
	}
	db.SetMaxOpenConns(maxOpenConnections)
	db.SetMaxIdleConns(maxOpenConnections)

	closeOnError := func(err error) (*sql.DB, error) {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(err, fmt.Errorf("close SQLite database: %w", closeErr))
		}
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return closeOnError(fmt.Errorf("ping SQLite database %q: %w", path, err))
	}
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		return closeOnError(fmt.Errorf("enable SQLite WAL journal mode: %w", err))
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migrations)
	if err != nil {
		return closeOnError(fmt.Errorf("prepare SQLite migrations: %w", err))
	}
	if _, err := provider.Up(ctx); err != nil {
		return closeOnError(fmt.Errorf("apply SQLite migrations: %w", err))
	}

	return db, nil
}
