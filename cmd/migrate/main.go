package main

import (
	"context"
	"log/slog"
	"os"
	"time"

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
		logger.Error("DATABASE_URL is required to run migrations")
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

	applied, err := database.RunMigrations(ctx, db, os.DirFS("migrations"))
	if err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations completed", "applied", applied)
}
