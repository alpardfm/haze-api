package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/alpardfm/haze-api/internal/config"
	"github.com/alpardfm/haze-api/internal/database"
	"github.com/alpardfm/haze-api/internal/worker"
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
		logger.Error("DATABASE_URL is required to run reminder worker")
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

	reminderWorker := worker.ReminderWorker{
		Store: worker.SQLReminderStore{DB: db},
	}

	result, err := reminderWorker.RunOnce(ctx)
	if err != nil {
		logger.Error("failed to run reminder worker", "error", err)
		os.Exit(1)
	}

	logger.Info("reminder worker completed", "scanned", result.Scanned, "sent", result.Sent, "skipped", result.Skipped)
}
