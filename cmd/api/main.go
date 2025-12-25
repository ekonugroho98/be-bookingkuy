package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ekonugroho98/be-bookingkuy/internal/auth"
	"github.com/ekonugroho98/be-bookingkuy/internal/search"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/config"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/health"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/middleware"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/server"
	"github.com/ekonugroho98/be-bookingkuy/internal/user"
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

	// Initialize event bus
	eb := eventbus.New()

	// Subscribe to auth events
	eb.Subscribe(context.Background(), eventbus.EventUserCreated, auth.HandleUserCreated)

	logger.Info("✅ Event bus initialized")

	// Initialize JWT manager
	jwtManager := jwt.NewManager(cfg.JWT.Secret)
	logger.Info("✅ JWT manager initialized")

	// Initialize repositories
	userRepo := user.NewRepository(database)
	authRepo := auth.NewRepository(database)

	// Initialize services
	userService := user.NewService(userRepo, eb)
	authService := auth.NewService(userRepo, authRepo, eb, jwtManager)
	searchService := search.NewService(search.NewRepository(database))

	// Initialize handlers
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(authService)
	searchHandler := search.NewHandler(searchService)

	// Setup router
	mux := http.NewServeMux()

	// Health check endpoints
	healthHandler := health.NewHandler(database)
	mux.HandleFunc("/health", healthHandler.Check)
	mux.HandleFunc("/health/ready", healthHandler.Ready)
	mux.HandleFunc("/health/live", healthHandler.Live)

	// Auth endpoints (public)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	// Search endpoints (public)
	mux.HandleFunc("POST /api/v1/search/hotels", searchHandler.SearchHotels)

	// User endpoints (protected)
	mux.HandleFunc("GET /api/v1/users/me", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(userHandler.GetProfile)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/users/me", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(userHandler.UpdateProfile)).ServeHTTP)

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
