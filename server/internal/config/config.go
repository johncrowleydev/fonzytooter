package config

import (
	"os"
	"time"
)

type Config struct {
	Address        string
	DatabasePath   string
	CurriculumPath string
	OpenRouter     OpenRouterConfig
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
		OpenRouter: OpenRouterConfig{
			APIKey:  os.Getenv("OPENROUTER_API_KEY"),
			Model:   os.Getenv("FONZYTOOTER_TUTOR_MODEL"),
			BaseURL: valueOrDefault("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
			Timeout: durationOrDefault("OPENROUTER_HTTP_TIMEOUT", 90*time.Second),
		},
	}
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
