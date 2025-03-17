package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bilbothegreedy/server-name-generator/internal/config"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Username,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Try to get admin user
	query := `SELECT COUNT(*) FROM users WHERE username = 'admin'`
	var count int
	err = db.QueryRowContext(context.Background(), query).Scan(&count)
	if err != nil {
		// Handle table doesn't exist
		if err.Error() == `pq: relation "users" does not exist` {
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
			_, err = db.ExecContext(context.Background(), createUsersTable)
			if err != nil {
				log.Fatalf("Failed to create users table: %v", err)
			}
			fmt.Println("Created users table")
		} else {
			log.Fatalf("Database error: %v", err)
		}
	}

	// Create admin user if it doesn't exist
	if count == 0 {
		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("adminpassword"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		// Create user
		now := time.Now().UTC()
		_, err = db.ExecContext(
			context.Background(),
			`INSERT INTO users (id, username, password, email, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New().String(),
			"admin",
			string(hashedPassword),
			"admin@example.com",
			"admin",
			now,
			now,
		)
		if err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}
		fmt.Println("Admin user created successfully")
	} else {
		fmt.Println("Admin user already exists")
	}
}
