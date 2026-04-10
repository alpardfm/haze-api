package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/alpardfm/haze-api/internal/auth"
	"github.com/alpardfm/haze-api/internal/config"
	"github.com/alpardfm/haze-api/internal/database"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if cfg.DatabaseURL == "" {
		logger.Error("DATABASE_URL is required to seed admin")
		os.Exit(1)
	}

	adminName := strings.TrimSpace(os.Getenv("ADMIN_NAME"))
	adminEmail := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	adminPhone := strings.TrimSpace(os.Getenv("ADMIN_PHONE"))
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminName == "" || adminEmail == "" || adminPhone == "" || adminPassword == "" {
		logger.Error("ADMIN_NAME, ADMIN_EMAIL, ADMIN_PHONE, and ADMIN_PASSWORD are required")
		os.Exit(1)
	}

	passwordHash, err := auth.HashPassword(adminPassword)
	if err != nil {
		logger.Error("failed to hash password", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	var adminID int64
	err = db.QueryRowContext(ctx, `
		INSERT INTO admins (name, email, phone, password_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			phone = EXCLUDED.phone,
			password_hash = EXCLUDED.password_hash,
			updated_at = now()
		RETURNING id
	`, adminName, adminEmail, adminPhone, passwordHash).Scan(&adminID)
	if err != nil {
		logger.Error("failed to seed admin", "error", err)
		os.Exit(1)
	}

	logger.Info("admin seeded", "id", adminID, "email", adminEmail)
}
