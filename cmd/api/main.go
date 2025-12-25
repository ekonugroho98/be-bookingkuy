package main

import (
	"fmt"
	"log"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/config"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

func main() {
	// Initialize logger
	logger.Init()

	// Load configuration
	cfg := config.Load()

	logger.Info("Starting Bookingkuy API Server...")
	logger.Info("Database: %s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	logger.Info("Redis: %s:%s", cfg.Redis.Host, cfg.Redis.Port)
	logger.Info("Server: %s:%s", cfg.Server.Host, cfg.Server.Port)

	// TODO: Initialize database connection
	// TODO: Initialize event bus
	// TODO: Initialize HTTP server
	// TODO: Register routes

	fmt.Println("✅ Bookingkuy API is ready!")
	log.Fatal("Server started (placeholder)")
}
