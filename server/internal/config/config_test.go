package config

import (
	"testing"
	"time"
)

func TestFromEnvUsesDefaultDatabasePath(t *testing.T) {
	t.Setenv("HELIX_ACADEMY_DB_PATH", "")

	cfg := FromEnv()

	if cfg.DatabasePath != "./data/helix-academy.db" {
		t.Fatalf("expected default database path, got %q", cfg.DatabasePath)
	}
}

func TestFromEnvConfiguresOpenRouterExplicitly(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "test-secret")
	t.Setenv("HELIX_ACADEMY_TUTOR_MODEL", "vendor/exact-model")
	t.Setenv("OPENROUTER_BASE_URL", "http://openrouter.test/v1")
	t.Setenv("OPENROUTER_HTTP_TIMEOUT", "12s")

	cfg := FromEnv()

	if !cfg.OpenRouter.Configured() || cfg.OpenRouter.APIKey != "test-secret" || cfg.OpenRouter.Model != "vendor/exact-model" {
		t.Fatalf("unexpected OpenRouter configuration: %#v", cfg.OpenRouter)
	}
	if cfg.OpenRouter.BaseURL != "http://openrouter.test/v1" || cfg.OpenRouter.Timeout != 12*time.Second {
		t.Fatalf("unexpected OpenRouter transport configuration: %#v", cfg.OpenRouter)
	}
}

func TestFromEnvLeavesOpenRouterUnconfiguredWithoutKeyOrModel(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("HELIX_ACADEMY_TUTOR_MODEL", "")
	t.Setenv("OPENROUTER_BASE_URL", "")
	t.Setenv("OPENROUTER_HTTP_TIMEOUT", "not-a-duration")

	cfg := FromEnv()

	if cfg.OpenRouter.Configured() {
		t.Fatalf("expected OpenRouter to remain disabled: %#v", cfg.OpenRouter)
	}
	if cfg.OpenRouter.BaseURL != "https://openrouter.ai/api/v1" || cfg.OpenRouter.Timeout != 90*time.Second {
		t.Fatalf("unexpected OpenRouter defaults: %#v", cfg.OpenRouter)
	}
}

func TestFromEnvUsesExplicitDatabasePath(t *testing.T) {
	t.Setenv("HELIX_ACADEMY_DB_PATH", `C:\learner-data\helix-academy.db`)

	cfg := FromEnv()

	if cfg.DatabasePath != `C:\learner-data\helix-academy.db` {
		t.Fatalf("expected explicit database path, got %q", cfg.DatabasePath)
	}
}

func TestFromEnvUsesSecureAuthenticationDefaults(t *testing.T) {
	t.Setenv("HELIX_ACADEMY_AUTH_USERNAME", "")
	t.Setenv("HELIX_ACADEMY_AUTH_PASSWORD", "")
	t.Setenv("HELIX_ACADEMY_AUTH_DISPLAY_NAME", "")
	t.Setenv("HELIX_ACADEMY_AUTH_SECURE_COOKIE", "")
	t.Setenv("HELIX_ACADEMY_AUTH_SESSION_TTL", "")

	cfg := FromEnv()

	if cfg.Authentication.Username != "" || cfg.Authentication.Password != "" {
		t.Fatal("expected authentication credentials to remain unconfigured")
	}
	if cfg.Authentication.DisplayName != "Owner" || !cfg.Authentication.SecureCookie || cfg.Authentication.SessionTTL != 24*time.Hour {
		t.Fatalf("unexpected authentication defaults: %#v", cfg.Authentication)
	}
}

func TestFromEnvConfiguresAuthenticationExplicitly(t *testing.T) {
	t.Setenv("HELIX_ACADEMY_AUTH_USERNAME", "learner@example.test")
	t.Setenv("HELIX_ACADEMY_AUTH_PASSWORD", "a-long-test-password")
	t.Setenv("HELIX_ACADEMY_AUTH_DISPLAY_NAME", "Learner")
	t.Setenv("HELIX_ACADEMY_AUTH_SECURE_COOKIE", "false")
	t.Setenv("HELIX_ACADEMY_AUTH_SESSION_TTL", "8h")

	cfg := FromEnv()

	if cfg.Authentication.Username != "learner@example.test" || cfg.Authentication.Password != "a-long-test-password" {
		t.Fatal("authentication credentials were not loaded from the environment")
	}
	if cfg.Authentication.DisplayName != "Learner" || cfg.Authentication.SecureCookie || cfg.Authentication.SessionTTL != 8*time.Hour {
		t.Fatalf("unexpected authentication settings: %#v", cfg.Authentication)
	}
}

func TestFromEnvDefaultsTutorAccessToDenied(t *testing.T) {
	t.Setenv("HELIX_ACADEMY_TUTOR_ENTITLED", "")
	t.Setenv("HELIX_ACADEMY_TUTOR_MONTHLY_TURN_LIMIT", "")

	cfg := FromEnv()

	if cfg.TutorAccess.Entitled || cfg.TutorAccess.MonthlyTurnLimit != 0 {
		t.Fatalf("unexpected default tutor access: %#v", cfg.TutorAccess)
	}
}

func TestFromEnvConfiguresTutorTurnAllowance(t *testing.T) {
	t.Setenv("HELIX_ACADEMY_TUTOR_ENTITLED", "true")
	t.Setenv("HELIX_ACADEMY_TUTOR_MONTHLY_TURN_LIMIT", "25")

	cfg := FromEnv()

	if !cfg.TutorAccess.Entitled || cfg.TutorAccess.MonthlyTurnLimit != 25 {
		t.Fatalf("unexpected tutor access: %#v", cfg.TutorAccess)
	}
}
