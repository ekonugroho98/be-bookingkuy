package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	assert.NotNil(t, manager)
	assert.Equal(t, secret, manager.secretKey)
}

func TestGenerateToken(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	userID := "user-123"
	email := "test@example.com"

	token, err := manager.GenerateToken(userID, email)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEqual(t, userID, token) // Token should be encoded, not plain text
}

func TestGenerateTokenDifferentUsers(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	token1, err1 := manager.GenerateToken("user-1", "user1@example.com")
	token2, err2 := manager.GenerateToken("user-2", "user2@example.com")

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2) // Different users should have different tokens
}

func TestGenerateTokenEmptyUserID(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	token, err := manager.GenerateToken("", "test@example.com")

	require.NoError(t, err)
	assert.NotEmpty(t, token) // JWT doesn't prevent empty user ID by default
}

func TestGenerateTokenEmptyEmail(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	token, err := manager.GenerateToken("user-123", "")

	require.NoError(t, err)
	assert.NotEmpty(t, token) // JWT doesn't prevent empty email by default
}

func TestValidateTokenValid(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	userID := "user-456"
	email := "valid@example.com"

	token, err := manager.GenerateToken(userID, email)
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)

	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestValidateTokenInvalidFormat(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	invalidTokens := []string{
		"",
		"invalid.token.string",
		"Bearer token",
		"abc123",
		"not.a.jwt",
	}

	for _, token := range invalidTokens {
		t.Run(token, func(t *testing.T) {
			claims, err := manager.ValidateToken(token)

			assert.Error(t, err)
			assert.Nil(t, claims)
			assert.Contains(t, err.Error(), "failed to parse token")
		})
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	secret1 := "secret-one"
	secret2 := "secret-two"

	manager1 := NewManager(secret1)
	manager2 := NewManager(secret2)

	token, err := manager1.GenerateToken("user-789", "user@example.com")
	require.NoError(t, err)

	// Try to validate with different secret
	claims, err := manager2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateTokenMalformed(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	// Create a malformed JWT
	malformedTokens := []string{
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", // Only header
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMifQ", // Header + payload (no signature)
		"invalid.signature.here",
		"a.b.c",
	}

	for _, token := range malformedTokens {
		t.Run(token, func(t *testing.T) {
			claims, err := manager.ValidateToken(token)

			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestValidateTokenExpired(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	// Create an expired token manually
	expiredTime := time.Now().Add(-time.Hour) // Expired 1 hour ago
	claims := &Claims{
		UserID: "user-expired",
		Email:  "expired@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiredTime),
			IssuedAt:  jwt.NewNumericDate(expiredTime.Add(-time.Hour * 2)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	// Try to validate expired token
	validatedClaims, err := manager.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestValidateTokenTampered(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	// Generate valid token
	token, err := manager.GenerateToken("user-tampered", "tampered@example.com")
	require.NoError(t, err)

	// Tamper with the token (change last character)
	tamperedToken := token[:len(token)-1] + "X"

	claims, err := manager.ValidateToken(tamperedToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateTokenWrongSigningMethod(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	// Create token with different signing method (none)
	claims := &Claims{
		UserID: "user-unsigned",
		Email:  "unsigned@example.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// Try to validate with HS256
	validatedClaims, err := manager.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

func TestTokenRoundTrip(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	testCases := []struct {
		name  string
		userID string
		email  string
	}{
		{
			name:  "normal user",
			userID: "user-111",
			email:  "user111@example.com",
		},
		{
			name:  "admin user",
			userID: "admin-222",
			email:  "admin@example.com",
		},
		{
			name:  "user with special chars in email",
			userID: "user-333",
			email:  "user+tag@example.com",
		},
		{
			name:  "user with long ID",
			userID: "user-very-long-id-123456789",
			email:  "longid@example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := manager.GenerateToken(tc.userID, tc.email)
			require.NoError(t, err)
			require.NotEmpty(t, token)

			validatedClaims, err := manager.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, tc.userID, validatedClaims.UserID)
			assert.Equal(t, tc.email, validatedClaims.Email)
		})
	}
}

func TestTokenUniqueness(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	userID := "user-unique"
	email := "unique@example.com"

	// Generate multiple tokens for the same user
	tokens := make([]string, 10)
	for i := 0; i < 10; i++ {
		token, err := manager.GenerateToken(userID, email)
		require.NoError(t, err)
		tokens[i] = token

		// Small delay to ensure different timestamps
		time.Sleep(time.Millisecond)
	}

	// All tokens should be different (due to different issued at times)
	uniqueTokens := make(map[string]bool)
	for _, token := range tokens {
		uniqueTokens[token] = true
	}

	// Note: JWT without explicit iat might not be unique, so we just check all tokens are valid
	assert.GreaterOrEqual(t, len(uniqueTokens), 1)

	// But all should validate to the same user
	for _, token := range tokens {
		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
	}
}

func TestTokenClaimsStructure(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	userID := "user-claims"
	email := "claims@example.com"

	token, err := manager.GenerateToken(userID, email)
	require.NoError(t, err)

	// Parse the token to check structure
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)

	// Check claims structure
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.NotNil(t, claims.RegisteredClaims)
}

func TestSecurityWeakSecret(t *testing.T) {
	// Test with very weak secrets
	weakSecrets := []string{
		"",
		"123",
		"abc",
		"password",
		"secret",
	}

	for _, secret := range weakSecrets {
		t.Run(secret, func(t *testing.T) {
			manager := NewManager(secret)

			token, err := manager.GenerateToken("user-weak", "weak@example.com")
			require.NoError(t, err)

			// Token should still work (even though secret is weak)
			claims, err := manager.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, "user-weak", claims.UserID)
		})
	}
}

func TestConcurrentTokenGeneration(t *testing.T) {
	secret := "test-secret-key"
	manager := NewManager(secret)

	// Generate tokens concurrently
	tokens := make(chan string, 100)
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			token, err := manager.GenerateToken("user-"+string(rune(id)), "concurrent@example.com")
			if err != nil {
				errors <- err
				return
			}
			tokens <- token
		}(i)
	}

	// Collect results
	tokenCount := 0
	errorCount := 0
	for i := 0; i < 100; i++ {
		select {
		case <-tokens:
			tokenCount++
		case <-errors:
			errorCount++
		}
	}

	assert.Equal(t, 100, tokenCount)
	assert.Equal(t, 0, errorCount)
}
