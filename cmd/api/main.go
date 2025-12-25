package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/config"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/health"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init()

	logger.Info("Starting Bookingkuy API Server...")
	logger.Info(fmt.Sprintf("Environment: %s", cfg.Environment))
	logger.Info(fmt.Sprintf("Database: %s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name))
	logger.Info(fmt.Sprintf("Redis: %s:%s", cfg.Redis.Host, cfg.Redis.Port))
	logger.Info(fmt.Sprintf("Server: %s:%s", cfg.Server.Host, cfg.Server.Port))

	// Initialize database connection
	database, err := db.New(context.Background(), cfg)
	if err != nil {
		logger.FatalWithErr(err, "Failed to connect to database")
		return
	}
	defer database.Close()

	logger.Info("✅ Database connected")

	// Initialize event bus (unused for now, will be used in Phase 2+)
	_ = eventbus.New()
	logger.Info("✅ Event bus initialized")

	// Setup router
	mux := http.NewServeMux()

	// Health check endpoints
	healthHandler := health.NewHandler(database)
	mux.HandleFunc("/health", healthHandler.Check)
	mux.HandleFunc("/health/ready", healthHandler.Ready)
	mux.HandleFunc("/health/live", healthHandler.Live)

	// TODO: Register API routes
	// mux.HandleFunc("/api/v1/auth/register", ...)
	// mux.HandleFunc("/api/v1/search/hotels", ...)
	// etc.

	logger.Info("✅ Routes registered")

	// Initialize and start HTTP server
	srv := server.New(cfg, mux)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			logger.FatalWithErr(err, "Server failed to start")
		}
	}()

	logger.Info("✅ Bookingkuy API is ready!")
	logger.Info(fmt.Sprintf("🚀 Server listening on %s:%s", cfg.Server.Host, cfg.Server.Port))

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx := context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		logger.ErrorWithErr(err, "Server shutdown error")
	}

	logger.Info("Server stopped")
}
