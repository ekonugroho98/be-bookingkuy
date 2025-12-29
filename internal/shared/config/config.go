package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Environment string
	Database    DatabaseConfig
	Redis       RedisConfig
	Server      ServerConfig
	JWT         JWTConfig
	Hotelbeds   HotelbedsConfig
	Midtrans    MidtransConfig
	SendGrid    SendGridConfig
	RabbitMQ    RabbitMQConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ServerConfig struct {
	Host string
	Port string
}

type JWTConfig struct {
	Secret     string
	Expiration string
}

type HotelbedsConfig struct {
	APIKey  string
	Secret  string
	BaseURL string
}

type MidtransConfig struct {
	MerchantID   string
	ClientKey    string
	ServerKey    string
	IsProduction bool
}

type SendGridConfig struct {
	APIKey    string
	FromEmail string
}

type RabbitMQConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	VHost          string
	ReconnectDelay time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	viper.SetEnvPrefix("BOOKINGKUY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	cfg := &Config{}

	// Bind environment variables explicitly to ensure mapping works
	viper.BindEnv("jwt.secret", "BOOKINGKUY_JWT_SECRET")
	viper.BindEnv("jwt.expiration", "BOOKINGKUY_JWT_EXPIRATION")
	viper.BindEnv("database.host", "BOOKINGKUY_DATABASE_HOST")
	viper.BindEnv("database.port", "BOOKINGKUY_DATABASE_PORT")
	viper.BindEnv("database.name", "BOOKINGKUY_DATABASE_NAME")
	viper.BindEnv("database.user", "BOOKINGKUY_DATABASE_USER")
	viper.BindEnv("database.password", "BOOKINGKUY_DATABASE_PASSWORD")
	viper.BindEnv("database.sslmode", "BOOKINGKUY_DATABASE_SSLMODE")
	viper.BindEnv("server.host", "BOOKINGKUY_SERVER_HOST")
	viper.BindEnv("server.port", "BOOKINGKUY_SERVER_PORT")
	viper.BindEnv("redis.host", "BOOKINGKUY_REDIS_HOST")
	viper.BindEnv("redis.port", "BOOKINGKUY_REDIS_PORT")
	viper.BindEnv("redis.password", "BOOKINGKUY_REDIS_PASSWORD")
	viper.BindEnv("redis.db", "BOOKINGKUY_REDIS_DB")
	viper.BindEnv("environment", "BOOKINGKUY_ENVIRONMENT")

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required configuration
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func setDefaults() {
	// Server
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")

	// Database
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.name", "bookingkuy_db")
	viper.SetDefault("database.user", "bookingkuy")
	viper.SetDefault("database.sslmode", "disable")

	// Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)

	// JWT
	viper.SetDefault("jwt.expiration", "24h")

	// Environment
	viper.SetDefault("environment", "development")

	// Hotelbeds
	viper.SetDefault("hotelbeds.baseurl", "https://api.hotelbeds.com")

	// Midtrans
	viper.SetDefault("midtrans.isproduction", false)

	// SendGrid
	viper.SetDefault("sendgrid.fromemail", "noreply@bookingkuy.com")

	// RabbitMQ
	viper.SetDefault("rabbitmq.host", "localhost")
	viper.SetDefault("rabbitmq.port", "5672")
	viper.SetDefault("rabbitmq.user", "guest")
	viper.SetDefault("rabbitmq.password", "guest")
	viper.SetDefault("rabbitmq.vhost", "/")
	viper.SetDefault("rabbitmq.reconnectdelay", "5s")
}

func validate(cfg *Config) error {
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if cfg.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	// For production, check for sensitive configurations
	if cfg.Environment == "production" {
		if cfg.Database.Password == "" {
			return fmt.Errorf("database password is required in production")
		}
		if cfg.JWT.Secret == "dev-secret-key" || cfg.JWT.Secret == "dev-secret-key-change-in-production" {
			return fmt.Errorf("JWT secret must be changed in production")
		}
	}

	return nil
}
