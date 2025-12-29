package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddlewareValidToken(t *testing.T) {
	secret := "test-secret"
	jwtManager := jwt.NewManager(secret)

	// Generate valid token
	userID := "user-123"
	email := "test@example.com"
	token, err := jwtManager.GenerateToken(userID, email)
	require.NoError(t, err)

	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if context has user info
		extractedUserID, ok := GetUserID(r.Context())
		require.True(t, ok)
		assert.Equal(t, userID, extractedUserID)

		extractedEmail, ok := GetEmail(r.Context())
		require.True(t, ok)
		assert.Equal(t, email, extractedEmail)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware
	middleware := AuthMiddleware(jwtManager)

	// Create request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	middleware(nextHandler).ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())
}

func TestAuthMiddlewareMissingAuthHeader(t *testing.T) {
	jwtManager := jwt.NewManager("test-secret")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("should not reach here"))
	})

	middleware := AuthMiddleware(jwtManager)

	req := httptest.NewRequest("GET", "/test", nil)
	// No Authorization header set

	rr := httptest.NewRecorder()

	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Authorization header required")
}

func TestAuthMiddlewareInvalidFormat(t *testing.T) {
	jwtManager := jwt.NewManager("test-secret")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtManager)

	testCases := []struct {
		name           string
		authHeader      string
		expectedMessage string
	}{
		{
			name:           "no Bearer prefix",
			authHeader:      "token-without-bearer",
			expectedMessage: "Invalid authorization header format",
		},
		{
			name:           "only Bearer",
			authHeader:      "Bearer",
			expectedMessage: "Invalid authorization header format",
		},
		{
			name:           "multiple parts",
			authHeader:      "Bearer token extra",
			expectedMessage: "Invalid authorization header format",
		},
		{
			name:           "wrong prefix",
			authHeader:      "Basic token",
			expectedMessage: "Invalid authorization header format",
		},
		{
			name:           "lowercase bearer",
			authHeader:      "bearer token",
			expectedMessage: "Invalid authorization header format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.authHeader)

			rr := httptest.NewRecorder()

			middleware(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedMessage)
		})
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	jwtManager := jwt.NewManager("test-secret")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtManager)

	invalidTokens := []struct {
		name  string
		token string
	}{
		{
			name:  "invalid string",
			token: "invalid-token-string",
		},
		{
			name:  "malformed JWT",
			token: "not.a.valid.jwt",
		},
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "random string",
			token: "abc123xyz789",
		},
	}

	for _, tc := range invalidTokens {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)

			rr := httptest.NewRecorder()

			middleware(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			assert.Contains(t, rr.Body.String(), "Invalid or expired token")
		})
	}
}

func TestAuthMiddlewareWrongSecret(t *testing.T) {
	secret1 := "secret-one"
	secret2 := "secret-two"

	jwtManager1 := jwt.NewManager(secret1)
	jwtManager2 := jwt.NewManager(secret2)

	// Generate token with secret1
	token, err := jwtManager1.GenerateToken("user-123", "test@example.com")
	require.NoError(t, err)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Validate with secret2
	middleware := AuthMiddleware(jwtManager2)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()

	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid or expired token")
}

func TestGetUserID(t *testing.T) {
	ctx := context.Background()

	// Test with user ID in context
	ctx = context.WithValue(ctx, UserIDKey, "user-456")

	userID, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "user-456", userID)
}

func TestGetUserIDNotInContext(t *testing.T) {
	ctx := context.Background()

	userID, ok := GetUserID(ctx)
	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestGetUserIDWrongType(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDKey, 12345) // Wrong type

	userID, ok := GetUserID(ctx)
	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestGetEmail(t *testing.T) {
	ctx := context.Background()

	// Test with email in context
	ctx = context.WithValue(ctx, EmailKey, "email@example.com")

	email, ok := GetEmail(ctx)
	assert.True(t, ok)
	assert.Equal(t, "email@example.com", email)
}

func TestGetEmailNotInContext(t *testing.T) {
	ctx := context.Background()

	email, ok := GetEmail(ctx)
	assert.False(t, ok)
	assert.Empty(t, email)
}

func TestAuthMiddlewareIntegration(t *testing.T) {
	secret := "integration-secret"
	jwtManager := jwt.NewManager(secret)

	// Test multiple successful authentications
	testCases := []struct {
		userID string
		email  string
	}{
		{"user-1", "user1@example.com"},
		{"user-2", "user2@example.com"},
		{"admin-1", "admin@example.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.userID, func(t *testing.T) {
			token, err := jwtManager.GenerateToken(tc.userID, tc.email)
			require.NoError(t, err)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				extractedUserID, ok := GetUserID(r.Context())
				require.True(t, ok)
				assert.Equal(t, tc.userID, extractedUserID)

				extractedEmail, ok := GetEmail(r.Context())
				require.True(t, ok)
				assert.Equal(t, tc.email, extractedEmail)

				w.WriteHeader(http.StatusOK)
			})

			middleware := AuthMiddleware(jwtManager)

			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()

			middleware(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}

func TestAuthMiddlewareChain(t *testing.T) {
	jwtManager := jwt.NewManager("chain-secret")

	token, err := jwtManager.GenerateToken("user-chain", "chain@example.com")
	require.NoError(t, err)

	// Create a chain of handlers
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("final"))
	})

	authMiddleware := AuthMiddleware(jwtManager)

	// Apply middleware
	handler := authMiddleware(finalHandler)

	req := httptest.NewRequest("GET", "/chain", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "final", rr.Body.String())
}

func TestAuthMiddlewarePreservesRequestMethod(t *testing.T) {
	jwtManager := jwt.NewManager("method-secret")

	token, err := jwtManager.GenerateToken("user-method", "method@example.com")
	require.NoError(t, err)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, method, r.Method)
				w.WriteHeader(http.StatusOK)
			})

			middleware := AuthMiddleware(jwtManager)

			req := httptest.NewRequest(method, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()

			middleware(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}

func TestAuthMiddlewarePreservesHeaders(t *testing.T) {
	jwtManager := jwt.NewManager("header-secret")

	token, err := jwtManager.GenerateToken("user-header", "header@example.com")
	require.NoError(t, err)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customHeader := r.Header.Get("X-Custom-Header")
		assert.Equal(t, "custom-value", customHeader)
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtManager)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Custom-Header", "custom-value")

	rr := httptest.NewRecorder()

	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddlewareContextValues(t *testing.T) {
	secret := "context-secret"
	jwtManager := jwt.NewManager(secret)

	userID := "context-user"
	email := "context@example.com"

	token, err := jwtManager.GenerateToken(userID, email)
	require.NoError(t, err)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all context values
		extractedUserID, ok := r.Context().Value(UserIDKey).(string)
		require.True(t, ok)
		assert.Equal(t, userID, extractedUserID)

		extractedEmail, ok := r.Context().Value(EmailKey).(string)
		require.True(t, ok)
		assert.Equal(t, email, extractedEmail)

		// Verify helper functions
		helperUserID, ok := GetUserID(r.Context())
		require.True(t, ok)
		assert.Equal(t, userID, helperUserID)

		helperEmail, ok := GetEmail(r.Context())
		require.True(t, ok)
		assert.Equal(t, email, helperEmail)

		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtManager)

	req := httptest.NewRequest("GET", "/context", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()

	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
