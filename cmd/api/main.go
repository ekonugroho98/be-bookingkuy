package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/admin"
	"github.com/ekonugroho98/be-bookingkuy/internal/auth"
	"github.com/ekonugroho98/be-bookingkuy/internal/booking"
	"github.com/ekonugroho98/be-bookingkuy/internal/midtrans"
	"github.com/ekonugroho98/be-bookingkuy/internal/payment"
	"github.com/ekonugroho98/be-bookingkuy/internal/pricing"
	"github.com/ekonugroho98/be-bookingkuy/internal/review"
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

	// Subscribe to booking events
	eb.Subscribe(context.Background(), eventbus.EventBookingCreated, booking.HandleBookingCreated)
	eb.Subscribe(context.Background(), eventbus.EventBookingPaid, booking.HandleBookingPaid)
	eb.Subscribe(context.Background(), eventbus.EventBookingConfirmed, booking.HandleBookingConfirmed)
	eb.Subscribe(context.Background(), eventbus.EventBookingCancelled, booking.HandleBookingCancelled)

	// Subscribe to payment events
	eb.Subscribe(context.Background(), eventbus.EventPaymentSuccess, payment.HandlePaymentSuccess)
	eb.Subscribe(context.Background(), eventbus.EventPaymentFailed, payment.HandlePaymentFailed)
	eb.Subscribe(context.Background(), eventbus.EventPaymentRefunded, payment.HandlePaymentRefunded)

	logger.Info("✅ Event bus initialized")

	// Initialize JWT manager
	jwtManager := jwt.NewManager(cfg.JWT.Secret)
	logger.Info("✅ JWT manager initialized")

	// Initialize Midtrans client
	midtransClient := midtrans.NewClient(midtrans.Config{
		ServerKey:    cfg.Midtrans.ServerKey,
		ClientKey:    cfg.Midtrans.ClientKey,
		MerchantID:   cfg.Midtrans.MerchantID,
		IsProduction: cfg.Midtrans.IsProduction,
	})
	logger.Info("✅ Midtrans client initialized")

	// Initialize repositories
	userRepo := user.NewRepository(database)
	authRepo := auth.NewRepository(database)

	// Initialize services
	userService := user.NewService(userRepo, eb)
	authService := auth.NewService(userRepo, authRepo, eb, jwtManager)
	searchService := search.NewService(search.NewRepository(database))
	pricingService := pricing.NewService()
	bookingService := booking.NewService(booking.NewRepository(database), eb, pricingService)
	paymentService := payment.NewServiceWithMidtrans(payment.NewRepository(database), eb, midtransClient)

	// Initialize admin service
	adminRepo := admin.NewRepository(database.Pool)
	adminService := admin.NewService(adminRepo, eb, cfg.JWT.Secret, 24*time.Hour)
	adminHandler := admin.NewHandler(adminService, cfg.JWT.Secret)

	// Initialize review service
	reviewRepo := review.NewRepository(database.Pool)
	reviewService := review.NewService(reviewRepo)
	reviewHandler := review.NewHandler(reviewService, cfg.JWT.Secret)

	// Initialize handlers
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(authService)
	searchHandler := search.NewHandler(searchService)
	bookingHandler := booking.NewHandler(bookingService)
	paymentHandler := payment.NewHandler(paymentService)

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

	// Booking endpoints (protected)
	mux.HandleFunc("POST /api/v1/bookings", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(bookingHandler.CreateBooking)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/bookings/{id}", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(bookingHandler.GetBooking)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/bookings/my", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(bookingHandler.GetMyBookings)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/bookings/{id}/cancel", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(bookingHandler.CancelBooking)).ServeHTTP)

	// Payment endpoints (protected + webhook)
	mux.HandleFunc("POST /api/v1/payments", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(paymentHandler.CreatePayment)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/payments/{id}", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(paymentHandler.GetPayment)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/payments/webhook", paymentHandler.HandleWebhook) // Public endpoint for webhooks

	// User endpoints (protected)
	mux.HandleFunc("GET /api/v1/users/me", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(userHandler.GetProfile)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/users/me", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(userHandler.UpdateProfile)).ServeHTTP)

	// Review endpoints (public + protected)
	mux.HandleFunc("GET /api/v1/reviews/hotel/", reviewHandler.GetHotelReviews)
	mux.HandleFunc("GET /api/v1/reviews/hotel/", reviewHandler.GetHotelStats)
	mux.HandleFunc("GET /api/v1/reviews/", reviewHandler.GetReviewByID)
	mux.HandleFunc("POST /api/v1/reviews", reviewHandler.CreateReview)
	mux.HandleFunc("PUT /api/v1/reviews/", reviewHandler.UpdateReview)
	mux.HandleFunc("DELETE /api/v1/reviews/", reviewHandler.DeleteReview)
	mux.HandleFunc("GET /api/v1/reviews/my-reviews", reviewHandler.GetMyReviews)
	mux.HandleFunc("POST /api/v1/reviews/", reviewHandler.ToggleHelpful)
	mux.HandleFunc("POST /api/v1/reviews/", reviewHandler.FlagReview)

	// Hotel partner endpoints (review responses)
	mux.HandleFunc("POST /api/v1/hotels/", reviewHandler.AddHotelResponse)
	mux.HandleFunc("PUT /api/v1/hotels/", reviewHandler.UpdateHotelResponse)
	mux.HandleFunc("DELETE /api/v1/hotels/", reviewHandler.DeleteHotelResponse)

	// Admin endpoints
	// Public admin endpoints
	mux.HandleFunc("POST /api/v1/admin/login", adminHandler.HandleLogin)

	// Protected admin endpoints
	adminAuth := adminHandler.AuthMiddleware
	mux.HandleFunc("POST /api/v1/admin/logout", adminAuth(adminHandler.HandleLogout))
	mux.HandleFunc("GET /api/v1/admin/me", adminAuth(adminHandler.HandleGetMe))
	mux.HandleFunc("GET /api/v1/admin/dashboard", adminAuth(adminHandler.HandleDashboard))

	// Admin management (requires admin:write permission)
	mux.HandleFunc("GET /api/v1/admin/admins", adminAuth(adminHandler.HandleListAdmins))
	mux.HandleFunc("POST /api/v1/admin/admins", adminAuth(adminHandler.HandleCreateAdmin))
	mux.HandleFunc("GET /api/v1/admin/admins/", adminAuth(adminHandler.HandleGetAdmin))
	mux.HandleFunc("PUT /api/v1/admin/admins/", adminAuth(adminHandler.HandleUpdateAdmin))
	mux.HandleFunc("DELETE /api/v1/admin/admins/", adminAuth(adminHandler.HandleDeleteAdmin))

	// User management (requires users:read/write permission)
	mux.HandleFunc("GET /api/v1/admin/users", adminAuth(adminHandler.HandleListUsers))
	mux.HandleFunc("GET /api/v1/admin/users/", adminAuth(adminHandler.HandleGetUser))
	mux.HandleFunc("PUT /api/v1/admin/users/", adminAuth(adminHandler.HandleUpdateUser))
	mux.HandleFunc("DELETE /api/v1/admin/users/", adminAuth(adminHandler.HandleDeleteUser))

	// Booking management (requires bookings:read/write permission)
	mux.HandleFunc("GET /api/v1/admin/bookings", adminAuth(adminHandler.HandleListBookings))
	mux.HandleFunc("GET /api/v1/admin/bookings/", adminAuth(adminHandler.HandleGetBooking))
	mux.HandleFunc("PUT /api/v1/admin/bookings/", adminAuth(adminHandler.HandleUpdateBooking))
	mux.HandleFunc("GET /api/v1/admin/bookings/stats", adminAuth(adminHandler.HandleBookingStats))

	// Provider management (requires providers:read/write permission)
	mux.HandleFunc("GET /api/v1/admin/providers", adminAuth(adminHandler.HandleListProviders))
	mux.HandleFunc("GET /api/v1/admin/providers/", adminAuth(adminHandler.HandleGetProvider))
	mux.HandleFunc("PUT /api/v1/admin/providers/", adminAuth(adminHandler.HandleUpdateProvider))

	// Analytics (requires analytics:read permission)
	mux.HandleFunc("GET /api/v1/admin/analytics/revenue", adminAuth(adminHandler.HandleRevenueStats))
	mux.HandleFunc("GET /api/v1/admin/analytics/users", adminAuth(adminHandler.HandleUserStats))
	mux.HandleFunc("GET /api/v1/admin/analytics/providers", adminAuth(adminHandler.HandleProviderStats))

	// Audit logs
	mux.HandleFunc("GET /api/v1/admin/audit-logs", adminAuth(adminHandler.HandleAuditLogs))

	// Review moderation (admin endpoints)
	mux.HandleFunc("GET /api/v1/admin/reviews/pending", adminAuth(reviewHandler.GetPendingReviews))
	mux.HandleFunc("GET /api/v1/admin/reviews/flagged", adminAuth(reviewHandler.GetFlaggedReviews))
	mux.HandleFunc("PUT /api/v1/admin/reviews/", adminAuth(reviewHandler.ModerateReview))
	mux.HandleFunc("GET /api/v1/admin/reviews/stats", adminAuth(reviewHandler.GetModerationStats))
	mux.HandleFunc("GET /api/v1/admin/reviews/analytics", adminAuth(reviewHandler.GetAnalytics))

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
