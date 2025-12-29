package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/stretchr/testify/require"
)

// GenerateTestToken creates a valid JWT token for testing
func GenerateTestToken(t *testing.T, userID, email string) string {
	t.Helper()
	jwtManager := jwt.NewManager("test-secret")
	token, err := jwtManager.GenerateToken(userID, email)
	require.NoError(t, err)
	return token
}

// GetTestContext returns a context with timeout for tests
func GetTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// GetTestUserID returns a test user ID
func GetTestUserID() string {
	return "test-user-123"
}

// GetTestUserEmail returns a test user email
func GetTestUserEmail() string {
	return "test@example.com"
}

// GetTestHotelID returns a test hotel ID
func GetTestHotelID() string {
	return "test-hotel-123"
}

// GetTestRoomID returns a test room ID
func GetTestRoomID() string {
	return "test-room-123"
}

// ParseDateHelper parses date string for tests
func ParseDateHelper(t *testing.T, dateStr string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", dateStr)
	require.NoError(t, err)
	return parsed
}
