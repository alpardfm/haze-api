package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	DatabaseURL string
	Timezone    *time.Location
}

func Load() (Config, error) {
	timezoneName := envOrDefault("APP_TIMEZONE", "Asia/Jakarta")
	timezone, err := time.LoadLocation(timezoneName)
	if err != nil {
		return Config{}, fmt.Errorf("load timezone %q: %w", timezoneName, err)
	}

	return Config{
		AppEnv:      envOrDefault("APP_ENV", "development"),
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Timezone:    timezone,
	}, nil
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
