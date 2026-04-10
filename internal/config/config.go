package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppEnv          string
	HTTPAddr        string
	DatabaseURL     string
	Timezone        *time.Location
	AuthTokenSecret string
	AuthTokenTTL    time.Duration
}

func Load() (Config, error) {
	timezoneName := envOrDefault("APP_TIMEZONE", "Asia/Jakarta")
	timezone, err := time.LoadLocation(timezoneName)
	if err != nil {
		return Config{}, fmt.Errorf("load timezone %q: %w", timezoneName, err)
	}

	authTokenTTL, err := parseDurationHours("AUTH_TOKEN_TTL_HOURS", 24)
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppEnv:          envOrDefault("APP_ENV", "development"),
		HTTPAddr:        envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		Timezone:        timezone,
		AuthTokenSecret: envOrDefault("AUTH_TOKEN_SECRET", "development-secret-change-me"),
		AuthTokenTTL:    authTokenTTL,
	}, nil
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func parseDurationHours(key string, fallbackHours int) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(fallbackHours) * time.Hour, nil
	}

	duration, err := time.ParseDuration(value + "h")
	if err != nil {
		return 0, fmt.Errorf("parse %s as hours: %w", key, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", key)
	}

	return duration, nil
}
