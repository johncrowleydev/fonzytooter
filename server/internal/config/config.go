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
	OpenRouter     OpenRouterConfig
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
		Address:        valueOrDefault("FONZYTOOTER_ADDR", ":8080"),
		DatabasePath:   valueOrDefault("FONZYTOOTER_DB_PATH", "./data/fonzytooter.db"),
		CurriculumPath: valueOrDefault("FONZYTOOTER_CURRICULUM_PATH", "../curriculum"),
		Authentication: AuthenticationConfig{
			Username:     os.Getenv("FONZYTOOTER_AUTH_USERNAME"),
			Password:     os.Getenv("FONZYTOOTER_AUTH_PASSWORD"),
			DisplayName:  valueOrDefault("FONZYTOOTER_AUTH_DISPLAY_NAME", "Owner"),
			SecureCookie: boolOrDefault("FONZYTOOTER_AUTH_SECURE_COOKIE", true),
			SessionTTL:   durationOrDefault("FONZYTOOTER_AUTH_SESSION_TTL", 24*time.Hour),
		},
		OpenRouter: OpenRouterConfig{
			APIKey:  os.Getenv("OPENROUTER_API_KEY"),
			Model:   os.Getenv("FONZYTOOTER_TUTOR_MODEL"),
			BaseURL: valueOrDefault("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
			Timeout: durationOrDefault("OPENROUTER_HTTP_TIMEOUT", 90*time.Second),
		},
	}
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
