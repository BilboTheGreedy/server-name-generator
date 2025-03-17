package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// InitializeAuthTables creates necessary authentication tables if they don't exist
func InitializeAuthTables(ctx context.Context, db *sql.DB) error {
	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		username VARCHAR(50) NOT NULL UNIQUE,
		password VARCHAR(100) NOT NULL,
		email VARCHAR(100) NOT NULL,
		role VARCHAR(20) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	)
	`
	if _, err := db.ExecContext(ctx, createUsersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create index on username for faster lookups
	createUsernameIndex := `
	CREATE INDEX IF NOT EXISTS idx_users_username
	ON users (username)
	`
	if _, err := db.ExecContext(ctx, createUsernameIndex); err != nil {
		return fmt.Errorf("failed to create username index: %w", err)
	}

	// Check if we need to create the default admin user
	var count int
	countQuery := `SELECT COUNT(*) FROM users WHERE role = 'admin'`
	err := db.QueryRowContext(ctx, countQuery).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for admin users: %w", err)
	}

	// If no admin users exist, create the default admin
	if count == 0 {
		// Get admin credentials from environment or use defaults
		adminUsername := os.Getenv("ADMIN_USERNAME")
		if adminUsername == "" {
			adminUsername = "admin"
		}

		adminPassword := os.Getenv("ADMIN_PASSWORD")
		if adminPassword == "" {
			adminPassword = "adminpassword" // Default password for development only
		}

		adminEmail := os.Getenv("ADMIN_EMAIL")
		if adminEmail == "" {
			adminEmail = "admin@example.com"
		}

		// Hash the admin password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}

		// Create the admin user
		now := time.Now().UTC()
		adminUser := models.User{
			ID:        uuid.New().String(),
			Username:  adminUsername,
			Password:  string(hashedPassword),
			Email:     adminEmail,
			Role:      models.RoleAdmin,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Insert admin user into the database
		insertQuery := `
		INSERT INTO users (id, username, password, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = db.ExecContext(ctx, insertQuery,
			adminUser.ID,
			adminUser.Username,
			adminUser.Password,
			adminUser.Email,
			adminUser.Role,
			adminUser.CreatedAt,
			adminUser.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		fmt.Printf("Default admin user created: %s\n", adminUsername)
	}

	return nil
}
