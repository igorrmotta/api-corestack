package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"

	"github.com/igorrmotta/api-corestack/services/golang/internal/config"
	"github.com/igorrmotta/api-corestack/services/golang/internal/repository"
	"github.com/igorrmotta/api-corestack/services/golang/internal/worker"
)

func main() {
	cfg := config.Load()

	// Configure slog
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database connection pool
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("worker connected to database")

	// Run River migrations
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		slog.Error("failed to create river migrator", "error", err)
		os.Exit(1)
	}
	res, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		slog.Error("failed to run river migrations", "error", err)
		os.Exit(1)
	}
	for _, v := range res.Versions {
		slog.Info("applied river migration", "version", v.Version)
	}

	// Initialize repositories
	taskRepo := repository.NewTaskRepo(pool)
	notifRepo := repository.NewNotificationRepo(pool)

	// Register River workers
	workers := river.NewWorkers()
	river.AddWorker(workers, worker.NewNotificationWorker(notifRepo))
	river.AddWorker(workers, worker.NewNotificationBatchWorker(notifRepo))
	river.AddWorker(workers, worker.NewImportWorker(taskRepo, notifRepo))

	// Initialize River client
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: cfg.RiverConcurrency},
		},
		Workers: workers,
	})
	if err != nil {
		slog.Error("failed to create river client", "error", err)
		os.Exit(1)
	}

	// Start River client
	if err := riverClient.Start(ctx); err != nil {
		slog.Error("failed to start river client", "error", err)
		os.Exit(1)
	}

	slog.Info("worker started", "concurrency", cfg.RiverConcurrency)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	slog.Info("worker shutting down", "signal", sig)
	cancel()

	riverClient.Stop(ctx)
	slog.Info("worker stopped")
}
