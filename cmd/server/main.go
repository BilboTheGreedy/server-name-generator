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
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// runMigrations handles database schema migrations
func runMigrations(dbURL string) error {
	// Construct migration URL
	migrationURL := fmt.Sprintf("file://migrations")

	// Create a new migrator
	m, err := migrate.New(migrationURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil {
		// It's okay if no change is needed
		if err == migrate.ErrNoChange {
			log.Println("No database migrations needed")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// Optional: Rollback function for downgrades
func rollbackMigrations(dbURL string) error {
	migrationURL := fmt.Sprintf("file://migrations")

	m, err := migrate.New(migrationURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Down(); err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	log.Println("Database migrations rolled back successfully")
	return nil
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := utils.NewLogger(cfg.LogLevel)
	logger.Info("Starting server name generator service")

	// Connect to database
	database, err := db.Connect(cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer database.Close()

	// Initialize router
	router := api.SetupRouter(cfg, database, logger)

	// Configure server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", "error", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give server max 5 seconds to finish requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exiting")
}
