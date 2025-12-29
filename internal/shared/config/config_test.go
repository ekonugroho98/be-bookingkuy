package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigWithDefaults(t *testing.T) {
	// Set minimal required environment variables
	os.Setenv("BOOKINGKUY_JWT_SECRET", "test-jwt-secret-for-testing")
	defer os.Unsetenv("BOOKINGKUY_JWT_SECRET")

	cfg, err := Load()

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check default values
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "bookingkuy_db", cfg.Database.Name)
	assert.Equal(t, "bookingkuy", cfg.Database.User)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, "6379", cfg.Redis.Port)
	assert.Equal(t, 0, cfg.Redis.DB)
	assert.Equal(t, "24h", cfg.JWT.Expiration)
	assert.Equal(t, "development", cfg.Environment)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"BOOKINGKUY_DATABASE_HOST":     "testhost",
		"BOOKINGKUY_DATABASE_PORT":     "5433",
		"BOOKINGKUY_DATABASE_NAME":     "testdb",
		"BOOKINGKUY_DATABASE_USER":     "testuser",
		"BOOKINGKUY_DATABASE_PASSWORD": "testpass",
		"BOOKINGKUY_DATABASE_SSLMODE":  "require",
		"BOOKINGKUY_JWT_SECRET":        "test-jwt-secret",
		"BOOKINGKUY_SERVER_HOST":       "127.0.0.1",
		"BOOKINGKUY_SERVER_PORT":       "9000",
		"BOOKINGKUY_REDIS_HOST":        "redishost",
		"BOOKINGKUY_REDIS_PORT":        "6380",
		"BOOKINGKUY_REDIS_DB":          "1",
		"BOOKINGKUY_ENVIRONMENT":       "production",
	}

	// Set all env vars
	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "testhost", cfg.Database.Host)
	assert.Equal(t, "5433", cfg.Database.Port)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, "test-jwt-secret", cfg.JWT.Secret)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "9000", cfg.Server.Port)
	assert.Equal(t, "redishost", cfg.Redis.Host)
	assert.Equal(t, "6380", cfg.Redis.Port)
	assert.Equal(t, 1, cfg.Redis.DB)
	assert.Equal(t, "production", cfg.Environment)
}

func TestLoadConfigMissingRequiredFields(t *testing.T) {
	// Don't set JWT_SECRET (required)
	// First unset it if it exists
	os.Unsetenv("BOOKINGKUY_JWT_SECRET")

	cfg, err := Load()

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT secret is required")
}

func TestLoadConfigProductionMissingPassword(t *testing.T) {
	envVars := map[string]string{
		"BOOKINGKUY_ENVIRONMENT":       "production",
		"BOOKINGKUY_JWT_SECRET":        "secure-production-secret",
		"BOOKINGKUY_DATABASE_HOST":     "localhost",
		"BOOKINGKUY_DATABASE_NAME":     "testdb",
		"BOOKINGKUY_DATABASE_USER":     "testuser",
		// Missing DATABASE_PASSWORD
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg, err := Load()

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "database password is required in production")
}

func TestLoadConfigProductionInsecureJWTSecret(t *testing.T) {
	envVars := map[string]string{
		"BOOKINGKUY_ENVIRONMENT":       "production",
		"BOOKINGKUY_JWT_SECRET":        "dev-secret-key-change-in-production",
		"BOOKINGKUY_DATABASE_HOST":     "localhost",
		"BOOKINGKUY_DATABASE_NAME":     "testdb",
		"BOOKINGKUY_DATABASE_USER":     "testuser",
		"BOOKINGKUY_DATABASE_PASSWORD": "prod-password",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg, err := Load()

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT secret must be changed in production")
}

func TestLoadConfigWithMidtrans(t *testing.T) {
	envVars := map[string]string{
		"BOOKINGKUY_JWT_SECRET":           "test-secret",
		"BOOKINGKUY_MIDTRANS_MERCHANTID":  "merchant123",
		"BOOKINGKUY_MIDTRANS_CLIENTKEY":   "client-key",
		"BOOKINGKUY_MIDTRANS_SERVERKEY":   "server-key",
		"BOOKINGKUY_MIDTRANS_ISPRODUCTION": "true",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "merchant123", cfg.Midtrans.MerchantID)
	assert.Equal(t, "client-key", cfg.Midtrans.ClientKey)
	assert.Equal(t, "server-key", cfg.Midtrans.ServerKey)
	assert.True(t, cfg.Midtrans.IsProduction)
}

func TestLoadConfigWithSendGrid(t *testing.T) {
	envVars := map[string]string{
		"BOOKINGKUY_JWT_SECRET":         "test-secret",
		"BOOKINGKUY_SENDGRID_APIKEY":    "sg-api-key",
		"BOOKINGKUY_SENDGRID_FROMEMAIL": "test@bookingkuy.com",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "sg-api-key", cfg.SendGrid.APIKey)
	assert.Equal(t, "test@bookingkuy.com", cfg.SendGrid.FromEmail)
}

func TestLoadConfigWithRabbitMQ(t *testing.T) {
	envVars := map[string]string{
		"BOOKINGKUY_JWT_SECRET":              "test-secret",
		"BOOKINGKUY_RABBITMQ_HOST":           "rabbitmq-host",
		"BOOKINGKUY_RABBITMQ_PORT":           "5673",
		"BOOKINGKUY_RABBITMQ_USER":           "rabbituser",
		"BOOKINGKUY_RABBITMQ_PASSWORD":       "rabbitpass",
		"BOOKINGKUY_RABBITMQ_VHOST":          "/testvhost",
		"BOOKINGKUY_RABBITMQ_RECONNECTDELAY": "10s",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "rabbitmq-host", cfg.RabbitMQ.Host)
	assert.Equal(t, "5673", cfg.RabbitMQ.Port)
	assert.Equal(t, "rabbituser", cfg.RabbitMQ.User)
	assert.Equal(t, "rabbitpass", cfg.RabbitMQ.Password)
	assert.Equal(t, "/testvhost", cfg.RabbitMQ.VHost)
	assert.Equal(t, 10*time.Second, cfg.RabbitMQ.ReconnectDelay)
}

func TestLoadConfigWithHotelbeds(t *testing.T) {
	envVars := map[string]string{
		"BOOKINGKUY_JWT_SECRET":        "test-secret",
		"BOOKINGKUY_HOTELBEDS_APIKEY":  "hb-api-key",
		"BOOKINGKUY_HOTELBEDS_SECRET":  "hb-secret",
		"BOOKINGKUY_HOTELBEDS_BASEURL": "https://test.hotelbeds.com",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "hb-api-key", cfg.Hotelbeds.APIKey)
	assert.Equal(t, "hb-secret", cfg.Hotelbeds.Secret)
	assert.Equal(t, "https://test.hotelbeds.com", cfg.Hotelbeds.BaseURL)
}
