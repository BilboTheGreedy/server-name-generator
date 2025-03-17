package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/bilbothegreedy/server-name-generator/internal/config"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// Connect opens the database connection, sets the pool parameters, tests the connection,
// and then runs migrations.
func Connect(cfg config.DatabaseConfig, logger *utils.Logger) (*sql.DB, error) {
	// Construct connection string in URL format
	urlQuery := url.Values{}
	if cfg.SSLMode != "" {
		urlQuery.Add("sslmode", cfg.SSLMode)
	}

	dbURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:     "/" + cfg.Name,
		RawQuery: urlQuery.Encode(),
	}

	// Open connection to database
	db, err := sql.Open("postgres", dbURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MaxConns / 2)
	db.SetConnMaxLifetime(time.Hour)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run database migrations
	if logger != nil {
		if err := runMigrations(dbURL.String(), logger); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return db, nil
}

// runMigrations locates the migrations folder, creates a migrator instance, and runs migrations.
func runMigrations(dbURL string, logger *utils.Logger) error {
	// Ensure the URL has the postgres:// scheme
	if !strings.HasPrefix(dbURL, "postgres://") && !strings.HasPrefix(dbURL, "postgresql://") {
		dbURL = fmt.Sprintf("postgres://%s", dbURL)
	}

	// Define possible migration paths in prioritized order
	possiblePaths := []string{
		// Try current working directory (often the project root if you run from there)
		func() string {
			wd, err := os.Getwd()
			if err == nil {
				return filepath.Join(wd, "migrations")
			}
			return ""
		}(),
		// Try one directory up (useful if the executable is in a subfolder)
		"../migrations",
		// Try executable directory
		func() string {
			execPath, err := os.Executable()
			if err == nil {
				return filepath.Join(filepath.Dir(execPath), "migrations")
			}
			return ""
		}(),
		// Try parent directory of executable
		func() string {
			execPath, err := os.Executable()
			if err == nil {
				parentDir := filepath.Dir(filepath.Dir(execPath))
				return filepath.Join(parentDir, "migrations")
			}
			return ""
		}(),
		// Fallback relative paths
		"./migrations",
		"../../migrations",
		// Hardcoded absolute path (adjust if necessary)
		"C:\\projects\\HostnameService\\server-name-generator\\migrations",
	}

	// Find the first existing migration path
	var migrationPath string
	for _, path := range possiblePaths {
		if path == "" {
			continue
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		// Check if directory exists and contains migration files
		if stat, err := os.Stat(absPath); err == nil && stat.IsDir() {
			files, err := os.ReadDir(absPath)
			if err == nil && len(files) > 0 {
				migrationPath = absPath
				break
			}
		}
	}

	// Validate migration path
	if migrationPath == "" {
		return fmt.Errorf("could not find migrations directory")
	}

	// Log the migration path chosen for debugging purposes
	if logger != nil {
		logger.Info("Using migration path", "path", migrationPath)
	}

	// Normalize path for URL (works on both Windows and Unix-like systems)
	// On Windows, using "file://" (two slashes) instead of "file:///" may resolve the error.
	migrationURL := "file://" + strings.ReplaceAll(filepath.ToSlash(migrationPath), "\\", "/")

	if logger != nil {
		logger.Info("Constructed migration URL", "url", migrationURL)
	}

	// Create migrator
	m, err := migrate.New(
		migrationURL,
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if logger != nil {
		logger.Info("Database migrations completed successfully", "path", migrationPath)
	}

	return nil
}

// initializeDatabase creates necessary tables if they don't exist.
func initializeDatabase(ctx context.Context, db *sql.DB) error {
	// Create sequences table
	createSequencesTable := `
	CREATE TABLE IF NOT EXISTS sequences (
		unit_code VARCHAR(10) NOT NULL,
		type VARCHAR(10) NOT NULL,
		provider VARCHAR(20) NOT NULL,
		region VARCHAR(10) NOT NULL,
		environment VARCHAR(10) NOT NULL,
		function VARCHAR(20) NOT NULL,
		current_value INTEGER NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		PRIMARY KEY (unit_code, type, provider, region, environment, function)
	)
	`
	if _, err := db.ExecContext(ctx, createSequencesTable); err != nil {
		return fmt.Errorf("failed to create sequences table: %w", err)
	}

	// Create reservations table
	createReservationsTable := `
	CREATE TABLE IF NOT EXISTS reservations (
		id UUID PRIMARY KEY,
		server_name VARCHAR(100) NOT NULL UNIQUE,
		unit_code VARCHAR(10) NOT NULL,
		type VARCHAR(10) NOT NULL,
		provider VARCHAR(20) NOT NULL,
		region VARCHAR(10) NOT NULL,
		environment VARCHAR(10) NOT NULL,
		function VARCHAR(20) NOT NULL,
		sequence_num INTEGER NOT NULL,
		status VARCHAR(20) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	)
	`
	if _, err := db.ExecContext(ctx, createReservationsTable); err != nil {
		return fmt.Errorf("failed to create reservations table: %w", err)
	}

	// Create index on status for faster lookups
	createStatusIndex := `
	CREATE INDEX IF NOT EXISTS idx_reservations_status
	ON reservations (status)
	`
	if _, err := db.ExecContext(ctx, createStatusIndex); err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
	}

	// Initialize authentication tables and default admin user
	if err := InitializeAuthTables(ctx, db); err != nil {
		return fmt.Errorf("failed to initialize authentication tables: %w", err)
	}

	// Initialize API keys tables
	if err := InitializeAPIKeysTables(ctx, db); err != nil {
		return fmt.Errorf("failed to initialize API keys tables: %w", err)
	}

	return nil
}
