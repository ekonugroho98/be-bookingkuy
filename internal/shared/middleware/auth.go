package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

const (
	UserIDKey = "user_id"
	EmailKey  = "email"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check if it's a Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				logger.ErrorWithErr(err, "Failed to validate token")
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Set user context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetEmail extracts email from context
func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}
