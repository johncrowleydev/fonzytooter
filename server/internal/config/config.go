package config

import "os"

type Config struct {
	Address      string
	DatabasePath string
}

func FromEnv() Config {
	return Config{
		Address:      valueOrDefault("FONZYTOOTER_ADDR", ":8080"),
		DatabasePath: valueOrDefault("FONZYTOOTER_DATABASE_PATH", "./data/fonzytooter.db"),
	}
}

func valueOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}
