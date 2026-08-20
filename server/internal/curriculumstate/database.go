package curriculumstate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const busyTimeoutMilliseconds = 5_000

func OpenReadOnly(ctx context.Context, path string) (*sql.DB, error) {
	return openExisting(ctx, path, "ro")
}

func OpenReadWrite(ctx context.Context, path string) (*sql.DB, error) {
	return openExisting(ctx, path, "rw")
}

func openExisting(ctx context.Context, path, mode string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("database path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database %q: %w", path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("open SQLite database %q: path is a directory", path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve SQLite database path %q: %w", path, err)
	}
	parameters := url.Values{
		"_busy_timeout": []string{fmt.Sprint(busyTimeoutMilliseconds)},
		"_foreign_keys": []string{"on"},
		"mode":          []string{mode},
	}
	uriPath := filepath.ToSlash(absolute)
	if filepath.VolumeName(absolute) != "" && !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	dsn := (&url.URL{Scheme: "file", Opaque: uriPath, RawQuery: parameters.Encode()}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database %q: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open SQLite database %q: %w", path, err)
	}
	return db, nil
}
