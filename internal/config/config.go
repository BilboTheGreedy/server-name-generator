package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// DatabaseConfig holds all database connection parameters
type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Name     string
	SSLMode  string
	MaxConns int
	Timeout  time.Duration
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string
	TokenDuration time.Duration
}

// Config holds all configuration for the application
type Config struct {
	Port     int
	LogLevel string
	Database DatabaseConfig
	Auth     AuthConfig
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Server configuration
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	logLevel := getEnv("LOG_LEVEL", "info")

	// Database configuration
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	dbMaxConns, err := strconv.Atoi(getEnv("DB_MAX_CONNECTIONS", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_CONNECTIONS: %w", err)
	}

	dbTimeout, err := time.ParseDuration(getEnv("DB_TIMEOUT", "10s"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_TIMEOUT: %w", err)
	}

	// Authentication configuration
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		// For development only - in production, this should be required
		jwtSecret = "server-name-generator-development-secret-key"
		fmt.Println("WARNING: Using default JWT secret. Set JWT_SECRET environment variable for production.")
	}

	tokenDurationStr := getEnv("TOKEN_DURATION", "24h")
	tokenDuration, err := time.ParseDuration(tokenDurationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid TOKEN_DURATION: %w", err)
	}

	return &Config{
		Port:     port,
		LogLevel: logLevel,
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			Username: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "server_names"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: dbMaxConns,
			Timeout:  dbTimeout,
		},
		Auth: AuthConfig{
			JWTSecret:     jwtSecret,
			TokenDuration: tokenDuration,
		},
	}, nil
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
