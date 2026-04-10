package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alpardfm/haze-api/internal/appointment"
	"github.com/alpardfm/haze-api/internal/auth"
	"github.com/alpardfm/haze-api/internal/config"
	"github.com/alpardfm/haze-api/internal/database"
	"github.com/alpardfm/haze-api/internal/health"
	"github.com/alpardfm/haze-api/internal/publicschedule"
	"github.com/alpardfm/haze-api/internal/shared/response"
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

	ctx := context.Background()
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}
	if db != nil {
		defer db.Close()
	}

	mux := http.NewServeMux()
	health.RegisterRoutes(mux, health.Handler{
		DB: db,
	})

	tokenManager := auth.NewTokenManager(cfg.AuthTokenSecret, cfg.AuthTokenTTL)
	var authStore auth.Store
	if db != nil {
		authStore = auth.SQLStore{DB: db}
	}
	auth.RegisterRoutes(mux, auth.Handler{
		Service: &auth.Service{
			Store:        authStore,
			TokenManager: tokenManager,
		},
	})

	var appointmentStore appointment.Store
	if db != nil {
		appointmentStore = appointment.SQLStore{DB: db}
	}
	appointmentHandler := appointment.Handler{
		Service: &appointment.Service{
			Store:    appointmentStore,
			Timezone: cfg.Timezone,
		},
	}
	mux.Handle("POST /appointments", auth.RequireAuth(tokenManager, http.HandlerFunc(appointmentHandler.Create)))
	mux.Handle("GET /appointments", auth.RequireAuth(tokenManager, http.HandlerFunc(appointmentHandler.List)))
	mux.Handle("GET /appointments/{id}", auth.RequireAuth(tokenManager, http.HandlerFunc(appointmentHandler.Detail)))
	mux.Handle("PUT /appointments/{id}", auth.RequireAuth(tokenManager, http.HandlerFunc(appointmentHandler.Update)))
	mux.Handle("PATCH /appointments/{id}/cancel", auth.RequireAuth(tokenManager, http.HandlerFunc(appointmentHandler.Cancel)))

	var publicScheduleStore publicschedule.Store
	if db != nil {
		publicScheduleStore = publicschedule.SQLStore{
			DB:       db,
			Timezone: cfg.Timezone.String(),
		}
	}
	publicScheduleHandler := publicschedule.Handler{
		Service: &publicschedule.Service{
			Store:    publicScheduleStore,
			Timezone: cfg.Timezone,
		},
	}
	mux.HandleFunc("GET /public/schedules", publicScheduleHandler.List)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusNotFound, response.Envelope{
			Success: false,
			Message: "route not found",
			Error: map[string]string{
				"path": r.URL.Path,
			},
		})
	})

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      requestLogger(logger, mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api server listening", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		logger.Error("server failed", "error", err)
		os.Exit(1)
	case sig := <-shutdownCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).String(),
		)
	})
}
