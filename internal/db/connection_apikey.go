package db

import (
	"context"
	"database/sql"
	"fmt"
)

// InitializeAPIKeysTables creates necessary API keys tables if they don't exist
func InitializeAPIKeysTables(ctx context.Context, db *sql.DB) error {
	// Create api_keys table
	createAPIKeysTable := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id UUID PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		key VARCHAR(100) NOT NULL UNIQUE,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		description TEXT,
		scopes TEXT,
		last_used TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE,
		is_active BOOLEAN NOT NULL DEFAULT TRUE
	)
	`
	if _, err := db.ExecContext(ctx, createAPIKeysTable); err != nil {
		return fmt.Errorf("failed to create api_keys table: %w", err)
	}

	// Create index on user_id for faster lookups
	createUserIDIndex := `
	CREATE INDEX IF NOT EXISTS idx_api_keys_user_id
	ON api_keys (user_id)
	`
	if _, err := db.ExecContext(ctx, createUserIDIndex); err != nil {
		return fmt.Errorf("failed to create user_id index: %w", err)
	}

	// Create index on key for faster lookups
	createKeyIndex := `
	CREATE INDEX IF NOT EXISTS idx_api_keys_key
	ON api_keys (key)
	`
	if _, err := db.ExecContext(ctx, createKeyIndex); err != nil {
		return fmt.Errorf("failed to create key index: %w", err)
	}

	return nil
}
