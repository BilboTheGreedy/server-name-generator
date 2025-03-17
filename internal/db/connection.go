package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bilbothegreedy/server-name-generator/internal/config"
	_ "github.com/lib/pq"
)

// Connect establishes a connection to the database
func Connect(cfg config.DatabaseConfig) (*sql.DB, error) {
	// Construct connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	// Open connection to database
	db, err := sql.Open("postgres", connStr)
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

	// Run initialization queries
	if err := initializeDatabase(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

// initializeDatabase creates necessary tables if they don't exist
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
