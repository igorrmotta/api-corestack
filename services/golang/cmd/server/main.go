package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/igorrmotta/api-corestack/services/golang/gen/comment/v1/commentv1connect"
	"github.com/igorrmotta/api-corestack/services/golang/gen/notification/v1/notificationv1connect"
	"github.com/igorrmotta/api-corestack/services/golang/gen/project/v1/projectv1connect"
	"github.com/igorrmotta/api-corestack/services/golang/gen/task/v1/taskv1connect"
	"github.com/igorrmotta/api-corestack/services/golang/gen/workspace/v1/workspacev1connect"
	"github.com/igorrmotta/api-corestack/services/golang/internal/config"
	"github.com/igorrmotta/api-corestack/services/golang/internal/handler"
	"github.com/igorrmotta/api-corestack/services/golang/internal/middleware"
	"github.com/igorrmotta/api-corestack/services/golang/internal/repository"
	"github.com/igorrmotta/api-corestack/services/golang/internal/service"
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
	slog.Info("connected to database")

	// Initialize repositories
	workspaceRepo := repository.NewWorkspaceRepo(pool)
	projectRepo := repository.NewProjectRepo(pool)
	taskRepo := repository.NewTaskRepo(pool)
	commentRepo := repository.NewCommentRepo(pool)
	notifRepo := repository.NewNotificationRepo(pool)

	// Initialize services
	workspaceSvc := service.NewWorkspaceService(workspaceRepo)
	projectSvc := service.NewProjectService(projectRepo)
	taskSvc := service.NewTaskService(taskRepo, notifRepo)
	commentSvc := service.NewCommentService(commentRepo)
	importSvc := service.NewImportService(taskRepo, notifRepo, cfg.RiverConcurrency, 100)

	// Initialize handlers
	workspaceHandler := handler.NewWorkspaceHandler(workspaceSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	taskHandler := handler.NewTaskHandler(taskSvc, importSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	notificationHandler := handler.NewNotificationHandler(notifRepo)

	// Connect RPC interceptors
	interceptors := connect.WithInterceptors(
		middleware.NewLoggingInterceptor(),
		middleware.NewRecoveryInterceptor(),
	)

	// Register Connect RPC routes
	mux := http.NewServeMux()

	path, h := workspacev1connect.NewWorkspaceServiceHandler(workspaceHandler, interceptors)
	mux.Handle(path, h)

	path, h = projectv1connect.NewProjectServiceHandler(projectHandler, interceptors)
	mux.Handle(path, h)

	path, h = taskv1connect.NewTaskServiceHandler(taskHandler, interceptors)
	mux.Handle(path, h)

	path, h = commentv1connect.NewCommentServiceHandler(commentHandler, interceptors)
	mux.Handle(path, h)

	path, h = notificationv1connect.NewNotificationServiceHandler(notificationHandler, interceptors)
	mux.Handle(path, h)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	addr := ":" + cfg.GRPCPort
	server := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-sigCh
	slog.Info("shutting down", "signal", sig)
	cancel()
	if err := server.Shutdown(context.Background()); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
