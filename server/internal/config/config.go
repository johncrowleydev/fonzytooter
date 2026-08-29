package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address        string
	DatabasePath   string
	CurriculumPath string
	Authentication AuthenticationConfig
	TutorAccess    TutorAccessConfig
	OpenRouter     OpenRouterConfig
}

type TutorAccessConfig struct {
	Entitled         bool
	MonthlyTurnLimit int
}

type AuthenticationConfig struct {
	Username     string
	Password     string
	DisplayName  string
	SecureCookie bool
	SessionTTL   time.Duration
}

type OpenRouterConfig struct {
	APIKey  string
	Model   string
	BaseURL string
	Timeout time.Duration
}

func (c OpenRouterConfig) Configured() bool {
	return c.APIKey != "" && c.Model != ""
}

func FromEnv() Config {
	return Config{
		Address:        valueOrDefault("HELIX_ACADEMY_ADDR", ":8080"),
		DatabasePath:   valueOrDefault("HELIX_ACADEMY_DB_PATH", "./data/helix-academy.db"),
		CurriculumPath: valueOrDefault("HELIX_ACADEMY_CURRICULUM_PATH", "../curriculum"),
		Authentication: AuthenticationConfig{
			Username:     os.Getenv("HELIX_ACADEMY_AUTH_USERNAME"),
			Password:     os.Getenv("HELIX_ACADEMY_AUTH_PASSWORD"),
			DisplayName:  valueOrDefault("HELIX_ACADEMY_AUTH_DISPLAY_NAME", "Owner"),
			SecureCookie: boolOrDefault("HELIX_ACADEMY_AUTH_SECURE_COOKIE", true),
			SessionTTL:   durationOrDefault("HELIX_ACADEMY_AUTH_SESSION_TTL", 24*time.Hour),
		},
		TutorAccess: TutorAccessConfig{
			Entitled:         boolOrDefault("HELIX_ACADEMY_TUTOR_ENTITLED", false),
			MonthlyTurnLimit: intOrDefault("HELIX_ACADEMY_TUTOR_MONTHLY_TURN_LIMIT", 0),
		},
		OpenRouter: OpenRouterConfig{
			APIKey:  os.Getenv("OPENROUTER_API_KEY"),
			Model:   os.Getenv("HELIX_ACADEMY_TUTOR_MODEL"),
			BaseURL: valueOrDefault("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
			Timeout: durationOrDefault("OPENROUTER_HTTP_TIMEOUT", 90*time.Second),
		},
	}
}

func intOrDefault(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func boolOrDefault(name string, fallback bool) bool {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func valueOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

func durationOrDefault(name string, fallback time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
