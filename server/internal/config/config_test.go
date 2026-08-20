package config

import "testing"

func TestFromEnvUsesDefaultDatabasePath(t *testing.T) {
	t.Setenv("FONZYTOOTER_DB_PATH", "")

	cfg := FromEnv()

	if cfg.DatabasePath != "./data/fonzytooter.db" {
		t.Fatalf("expected default database path, got %q", cfg.DatabasePath)
	}
}

func TestFromEnvUsesExplicitDatabasePath(t *testing.T) {
	t.Setenv("FONZYTOOTER_DB_PATH", `C:\learner-data\fonzytooter.db`)

	cfg := FromEnv()

	if cfg.DatabasePath != `C:\learner-data\fonzytooter.db` {
		t.Fatalf("expected explicit database path, got %q", cfg.DatabasePath)
	}
}
