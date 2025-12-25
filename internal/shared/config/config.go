package config

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   ServerConfig
	JWT      JWTConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type ServerConfig struct {
	Host string
	Port string
}

type JWTConfig struct {
	Secret     string
	Expiration string
}

// Load loads configuration from environment
func Load() *Config {
	// TODO: Implement environment variable loading
	return &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			Name:     "bookingkuy_db",
			User:     "bookingkuy",
			Password: "bookingkuy_dev_password",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
		},
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: "8080",
		},
		JWT: JWTConfig{
			Secret:     "dev-secret-key",
			Expiration: "24h",
		},
	}
}
