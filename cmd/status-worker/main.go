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
		logger.Error("DATABASE_URL is required to run status worker")
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

	statusWorker := worker.StatusWorker{
		Store: worker.SQLStatusStore{DB: db},
	}

	result, err := statusWorker.RunOnce(ctx)
	if err != nil {
		logger.Error("failed to run status worker", "error", err)
		os.Exit(1)
	}

	logger.Info("status worker completed", "marked_on_going", result.MarkedOnGoing, "marked_done", result.MarkedDone)
}
