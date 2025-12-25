package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/config"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
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
	logger.Info("Environment: %s", cfg.Environment)
	logger.Info("Database: %s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	logger.Info("Redis: %s:%s", cfg.Redis.Host, cfg.Redis.Port)
	logger.Info("Server: %s:%s", cfg.Server.Host, cfg.Server.Port)

	// TODO: Initialize database connection
	// TODO: Initialize event bus
	// TODO: Initialize HTTP server
	// TODO: Register routes

	logger.Info("✅ Bookingkuy API is ready!")
	log.Fatal("Server started (placeholder - TODO: Implement HTTP server)")
}
