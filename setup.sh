#!/bin/bash
# Server Name Generator Project Setup Script

set -e  # Exit on any error

# Project name (used for go module)
PROJECT_NAME="github.com/bilbothegreedy/server-name-generator"

# Base directory (current directory)
BASE_DIR=$(pwd)/server-name-generator

# Print step information
step() {
  echo ""
  echo "🔷 $1"
  echo "------------------------------------------------------------"
}

# Create a directory if it doesn't exist
create_dir() {
  if [ ! -d "$1" ]; then
    mkdir -p "$1"
    echo "Created directory: $1"
  fi
}

# Create a file with content
create_file() {
  file_path="$1"
  file_dir=$(dirname "$file_path")
  create_dir "$file_dir"
  
  touch "$file_path"
  echo "Created file: $file_path"
}

# Initialize the project
step "Creating project structure"

# Create base project directory
create_dir "$BASE_DIR"
cd "$BASE_DIR"

# Create main directories
create_dir "cmd/server"
create_dir "internal/api/handlers"
create_dir "internal/api/middleware"
create_dir "internal/config"
create_dir "internal/db"
create_dir "internal/models"
create_dir "internal/services"
create_dir "internal/session"
create_dir "internal/templates/layouts"
create_dir "internal/templates/partials"
create_dir "internal/templates/admin"
create_dir "internal/templates/auth"
create_dir "internal/utils"
create_dir "internal/errors"
create_dir "migrations"
create_dir "static/css"
create_dir "static/js"
create_dir "static/img"

# Initialize go module
step "Initializing Go module"
go mod init $PROJECT_NAME
echo "Initialized Go module: $PROJECT_NAME"

# Create essential files
step "Creating essential files"

# Main file
cat > cmd/server/main.go << 'EOL'
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bilbothegreedy/server-name-generator/internal/api"
	"github.com/bilbothegreedy/server-name-generator/internal/config"
	"github.com/bilbothegreedy/server-name-generator/internal/db"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

func main() {
	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger.
	logger := utils.NewLogger(cfg.LogLevel)
	logger.Info("Starting server name generator service")

	// Capture application start time.
	startTime := time.Now()

	// Connect to database (this runs migrations as needed).
	database, err := db.Connect(cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer database.Close()

	// Initialize router
	router := api.SetupRouter(cfg, database, logger, startTime)

	// Configure HTTP server.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine.
	go func() {
		logger.Info("Server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", "error", err)
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exiting")
}
EOL

# Config
cat > internal/config/config.go << 'EOL'
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
	Port        int
	LogLevel    string
	Environment string
	Database    DatabaseConfig
	Auth        AuthConfig
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
	environment := getEnv("ENVIRONMENT", "development")

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
		Port:        port,
		LogLevel:    logLevel,
		Environment: environment,
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
EOL

# Logger
cat > internal/utils/logger.go << 'EOL'
package utils

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Logger is a wrapper around slog.Logger with predefined methods
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new logger with the specified level
func NewLogger(level string) *Logger {
	// Set log level based on configuration
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create logger options
	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format timestamps as RFC3339
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.Attr{
						Key:   slog.TimeKey,
						Value: slog.StringValue(t.Format(time.RFC3339)),
					}
				}
			}
			return a
		},
	}

	// Create JSON handler
	handler := slog.NewJSONHandler(os.Stdout, opts)

	// Create the logger
	logger := slog.New(handler)

	return &Logger{logger}
}

// WithContext creates a new logger with additional context
func (l *Logger) WithContext(ctx ...any) *Logger {
	return &Logger{l.Logger.With(ctx...)}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...any) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
}
EOL

# Session Manager
cat > internal/session/session.go << 'EOL'
package session

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
)

// SessionManager is a wrapper around scs.SessionManager
type SessionManager struct {
	*scs.SessionManager
}

// NewSessionManager creates a new session manager
func NewSessionManager(db *sql.DB) *SessionManager {
	// Create a new session manager
	sessionManager := scs.New()
	
	// Configure the session lifetime
	sessionManager.Lifetime = 24 * time.Hour
	
	// Use PostgreSQL as the session store
	sessionManager.Store = postgresstore.New(db)
	
	// Configure cookie security settings
	sessionManager.Cookie.Secure = true    // Set to true in production
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	
	return &SessionManager{
		SessionManager: sessionManager,
	}
}

// GetUserID gets the user ID from the session
func (sm *SessionManager) GetUserID(r *http.Request) string {
	return sm.GetString(r.Context(), "userID")
}

// SetUserID sets the user ID in the session
func (sm *SessionManager) SetUserID(r *http.Request, w http.ResponseWriter, userID string) {
	sm.Put(r.Context(), "userID", userID)
}

// GetUserRole gets the user role from the session
func (sm *SessionManager) GetUserRole(r *http.Request) string {
	return sm.GetString(r.Context(), "userRole")
}

// SetUserRole sets the user role in the session
func (sm *SessionManager) SetUserRole(r *http.Request, w http.ResponseWriter, role string) {
	sm.Put(r.Context(), "userRole", role)
}

// SetFlash sets a flash message
func (sm *SessionManager) SetFlash(r *http.Request, w http.ResponseWriter, message string, messageType string) {
	sm.Put(r.Context(), "flash", message)
	sm.Put(r.Context(), "flashType", messageType)
}

// GetFlash gets and clears the flash message
func (sm *SessionManager) GetFlash(r *http.Request) (string, string) {
	message := sm.PopString(r.Context(), "flash")
	messageType := sm.PopString(r.Context(), "flashType")
	return message, messageType
}

// IsAuthenticated checks if the user is authenticated
func (sm *SessionManager) IsAuthenticated(r *http.Request) bool {
	return sm.GetString(r.Context(), "userID") != ""
}

// IsAdmin checks if the user is an admin
func (sm *SessionManager) IsAdmin(r *http.Request) bool {
	return sm.GetString(r.Context(), "userRole") == "admin"
}

// Logout clears the session
func (sm *SessionManager) Logout(r *http.Request, w http.ResponseWriter) {
	sm.Destroy(r.Context())
}
EOL

# Reservation model
cat > internal/models/reservation.go << 'EOL'
package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Reservation statuses
const (
	StatusReserved  = "reserved"
	StatusCommitted = "committed"
)

// Reservation represents a server name reservation in the database
type Reservation struct {
	ID          string    `json:"id"`
	ServerName  string    `json:"serverName"`
	UnitCode    string    `json:"unitCode"`
	Type        string    `json:"type"`
	Provider    string    `json:"provider"`
	Region      string    `json:"region"`
	Environment string    `json:"environment"`
	Function    string    `json:"function"`
	SequenceNum int       `json:"sequenceNum"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ReservationPayload represents the request payload for reserving a server name
type ReservationPayload struct {
	UnitCode    string `json:"unitCode,omitempty"`
	Type        string `json:"type,omitempty"`
	Provider    string `json:"provider,omitempty"`
	Region      string `json:"region,omitempty"`
	Environment string `json:"environment,omitempty"`
	Function    string `json:"function,omitempty"`
}

// ReservationResponse is the API response for reservation operations
type ReservationResponse struct {
	ReservationID string `json:"reservationId"`
	ServerName    string `json:"serverName"`
}

// CommitPayload represents the request payload for committing a reservation
type CommitPayload struct {
	ReservationID string `json:"reservationId"`
}

// ReleasePayload represents the request payload for releasing a reservation
type ReleasePayload struct {
	ReservationID string `json:"reservationId"`
}

// ReservationModel handles database operations for reservations
type ReservationModel struct {
	DB *sql.DB
}

// NewReservationModel creates a new reservation model
func NewReservationModel(db *sql.DB) *ReservationModel {
	return &ReservationModel{DB: db}
}

// Create inserts a new reservation into the database
func (m *ReservationModel) Create(ctx context.Context, tx *sql.Tx, r *Reservation) error {
	// Implementation details
	return nil
}

// GetByID retrieves a reservation by its ID
func (m *ReservationModel) GetByID(ctx context.Context, id string) (*Reservation, error) {
	// Implementation details
	return nil, nil
}

// IsServerNameUnique checks if a server name is already in use
func (m *ReservationModel) IsServerNameUnique(ctx context.Context, tx *sql.Tx, serverName string) (bool, error) {
	// Implementation details
	return true, nil
}

// FindLatestSequenceNumber finds the latest sequence number for a similar name pattern
func (m *ReservationModel) FindLatestSequenceNumber(ctx context.Context, tx *sql.Tx, pattern string) (int, error) {
	// Implementation details
	return 0, nil
}

// UpdateStatus updates the status of a reservation
func (m *ReservationModel) UpdateStatus(ctx context.Context, tx *sql.Tx, id, status string) error {
	// Implementation details
	return nil
}

// Delete deletes a reservation by ID (works for any status)
func (m *ReservationModel) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	// Implementation details
	return nil
}

// Release changes a committed reservation back to reserved
func (m *ReservationModel) Release(ctx context.Context, tx *sql.Tx, id string) error {
	// Implementation details
	return nil
}
EOL

# Create .env file
cat > .env << 'EOL'
# Server Configuration
PORT=8080
LOG_LEVEL=info
ENVIRONMENT=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres  # Use the default postgres user
DB_PASSWORD=postgres
DB_NAME=server_names
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=10
DB_TIMEOUT=10s

# Authentication Settings
JWT_SECRET=long_random_secret_key_min_32_chars
TOKEN_DURATION=24h

# Admin Credentials (change in production!)
ADMIN_USERNAME=admin
ADMIN_PASSWORD=adminpassword
ADMIN_EMAIL=admin@example.com
EOL

# Create docker-compose.yml
cat > docker-compose.yml << 'EOL'
version: '3.8'

services:
  # PostgreSQL database service
  database:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=server_names
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
EOL

# Create go.mod dependencies
step "Installing Go dependencies"
go get -u github.com/alexedwards/scs/v2
go get -u github.com/alexedwards/scs/postgresstore
go get -u github.com/go-chi/chi/v5
go get -u github.com/gorilla/csrf
go get -u github.com/justinas/nosurf
go get -u github.com/lib/pq
go get -u github.com/joho/godotenv
go get -u github.com/google/uuid
go get -u golang.org/x/crypto/bcrypt

# Create basic template files
step "Creating template files"

# Base template layout
create_file "internal/templates/layouts/base.tmpl"
create_file "internal/templates/layouts/admin.tmpl"
create_file "internal/templates/partials/header.tmpl"
create_file "internal/templates/partials/footer.tmpl"
create_file "internal/templates/partials/nav.tmpl"
create_file "internal/templates/admin/dashboard.tmpl"
create_file "internal/templates/admin/generate.tmpl"
create_file "internal/templates/admin/manage.tmpl"
create_file "internal/templates/auth/login.tmpl"

# Create more essential files
create_file "internal/handlers/render.go"
create_file "internal/handlers/admin.go"
create_file "internal/handlers/auth.go"
create_file "internal/handlers/api.go"
create_file "internal/utils/json.go"
create_file "internal/middleware/auth.go"
create_file "internal/models/user.go"
create_file "internal/models/sequence.go"
create_file "internal/api/routes.go"
create_file "internal/db/connection.go"

# Create migrations directory
step "Creating database migrations"
create_file "migrations/001_initial_schema.up.sql"
create_file "migrations/001_initial_schema.down.sql"

# Create main CSS and JS files
create_file "static/css/main.css"
create_file "static/js/app.js"

# Create favicon
create_file "static/favicon.ico"

# Create README
cat > README.md << 'EOL'
# Server Name Generator

A Go-based application for generating and managing server names.

## Getting Started

### Prerequisites
- Go 1.21 or later
- PostgreSQL database
- Docker and Docker Compose (optional)

### Setup and Run

1. Clone the repository
```bash
git clone https://github.com/bilbothegreedy/server-name-generator.git
cd server-name-generator
```

2. Start the database (using Docker Compose)
```bash
docker-compose up -d database
```

3. Run the application
```bash
go run cmd/server/main.go
```

4. Access the application at http://localhost:8080

### Environment Variables

See the `.env` file for configuration options.

## Features

- Server name generation and management
- Admin dashboard
- User authentication
- API access
EOL

step "Project creation complete!"
echo "Your Go-based Server Name Generator project has been set up."
echo ""
echo "Start the database with: docker-compose up -d database"
echo "Run the application with: go run cmd/server/main.go"
echo "Access the web interface at: http://localhost:8080"