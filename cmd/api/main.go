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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgdevgo/subscription-api/config"
	_ "github.com/kgdevgo/subscription-api/docs"
	deliveryHTTP "github.com/kgdevgo/subscription-api/internal/delivery/http"
	"github.com/kgdevgo/subscription-api/internal/repository/postgres"
	"github.com/kgdevgo/subscription-api/internal/usecase"
)

// @title           Subscription Aggregation API
// @version         1.0
// @description     REST-сервис для агрегации данных об онлайн подписках пользователей.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	// Initialize logger
	var logHandler slog.Handler
	cfg := config.MustLoad()

	if cfg.Env == "prod" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	slog.Info("starting subscription aggregation service", slog.String("env", cfg.Env))

	// Initialize database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.PostgreSQL.DSN)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	// Ping database
	if err := dbPool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", slog.String("err", err.Error()))
		os.Exit(1)
	}
	slog.Info("successfully connected to database")

	// Initialize dependencies
	repo := postgres.NewSubscriptionRepo(dbPool)
	uCase := usecase.NewSubscriptionUseCase(repo)
	router := deliveryHTTP.NewRouter(uCase)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPServer.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// Start HTTP server
	go func() {
		slog.Info("http server is running", slog.String("port", cfg.HTTPServer.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown http server cleanly", slog.String("err", err.Error()))
		os.Exit(1)
	}

	slog.Info("server exited successfully")
}
