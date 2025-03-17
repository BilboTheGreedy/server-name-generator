package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API key in the system
type APIKey struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Key         string     `json:"key,omitempty"` // Only included when first created
	UserID      string     `json:"userId"`
	Description string     `json:"description"`
	Scopes      []string   `json:"scopes"`
	LastUsed    *time.Time `json:"lastUsed,omitempty"` // Pointer to handle NULL values
	CreatedAt   time.Time  `json:"createdAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"` // Pointer to handle NULL values
	IsActive    bool       `json:"isActive"`
}

// APIKeyModel handles database operations for API keys
type APIKeyModel struct {
	DB *sql.DB
}

// NewAPIKeyModel creates a new API key model
func NewAPIKeyModel(db *sql.DB) *APIKeyModel {
	return &APIKeyModel{DB: db}
}

// Create inserts a new API key into the database
func (m *APIKeyModel) Create(ctx context.Context, tx *sql.Tx, apiKey *APIKey) error {
	query := `
		INSERT INTO api_keys (
			id, name, key, user_id, description, scopes, created_at, expires_at, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	// Create a transaction if one isn't provided
	var err error
	if tx == nil {
		tx, err = m.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()
	}

	// Convert scopes array to a comma-separated string
	scopesStr := ""
	if len(apiKey.Scopes) > 0 {
		for i, scope := range apiKey.Scopes {
			if i > 0 {
				scopesStr += ","
			}
			scopesStr += scope
		}
	}

	_, err = tx.ExecContext(
		ctx,
		query,
		apiKey.ID,
		apiKey.Name,
		apiKey.Key,
		apiKey.UserID,
		apiKey.Description,
		scopesStr,
		apiKey.CreatedAt,
		apiKey.ExpiresAt,
		apiKey.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	return nil
}

// GetByID retrieves an API key by its ID
func (m *APIKeyModel) GetByID(ctx context.Context, id string) (*APIKey, error) {
	query := `
		SELECT id, name, user_id, description, scopes, last_used, created_at, expires_at, is_active
		FROM api_keys
		WHERE id = $1
	`

	var scopesStr string
	var lastUsed sql.NullTime
	var expiresAt sql.NullTime
	apiKey := &APIKey{}
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&apiKey.ID,
		&apiKey.Name,
		&apiKey.UserID,
		&apiKey.Description,
		&scopesStr,
		&lastUsed,
		&apiKey.CreatedAt,
		&expiresAt,
		&apiKey.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by ID: %w", err)
	}

	// Convert scopes string to array
	if scopesStr != "" {
		apiKey.Scopes = splitScopesString(scopesStr)
	}

	// Set LastUsed only if the database value is not NULL
	if lastUsed.Valid {
		apiKey.LastUsed = &lastUsed.Time
	}

	// Set ExpiresAt only if the database value is not NULL
	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}

	return apiKey, nil
}

// GetByKey retrieves an API key by its key value
func (m *APIKeyModel) GetByKey(ctx context.Context, key string) (*APIKey, error) {
	query := `
		SELECT id, name, user_id, description, scopes, last_used, created_at, expires_at, is_active
		FROM api_keys
		WHERE key = $1 AND is_active = true AND (expires_at IS NULL OR expires_at > NOW())
	`

	var scopesStr string
	var lastUsed sql.NullTime
	var expiresAt sql.NullTime
	apiKey := &APIKey{}
	err := m.DB.QueryRowContext(ctx, query, key).Scan(
		&apiKey.ID,
		&apiKey.Name,
		&apiKey.UserID,
		&apiKey.Description,
		&scopesStr,
		&lastUsed,
		&apiKey.CreatedAt,
		&expiresAt,
		&apiKey.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by key: %w", err)
	}

	// Convert scopes string to array
	if scopesStr != "" {
		apiKey.Scopes = splitScopesString(scopesStr)
	}

	// Set LastUsed only if the database value is not NULL
	if lastUsed.Valid {
		apiKey.LastUsed = &lastUsed.Time
	}

	// Set ExpiresAt only if the database value is not NULL
	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}

	// Update last used timestamp
	_, err = m.DB.ExecContext(
		ctx,
		`UPDATE api_keys SET last_used = NOW() WHERE id = $1`,
		apiKey.ID,
	)
	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("Failed to update last_used for API key %s: %v\n", apiKey.ID, err)
	}

	return apiKey, nil
}

// GetByUserID retrieves all API keys for a user
func (m *APIKeyModel) GetByUserID(ctx context.Context, userID string) ([]*APIKey, error) {
	query := `
		SELECT id, name, user_id, description, scopes, last_used, created_at, expires_at, is_active
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*APIKey
	for rows.Next() {
		var scopesStr string
		var lastUsed sql.NullTime
		var expiresAt sql.NullTime
		apiKey := &APIKey{}
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.Name,
			&apiKey.UserID,
			&apiKey.Description,
			&scopesStr,
			&lastUsed,
			&apiKey.CreatedAt,
			&expiresAt,
			&apiKey.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key row: %w", err)
		}

		// Convert scopes string to array
		if scopesStr != "" {
			apiKey.Scopes = splitScopesString(scopesStr)
		}

		// Set LastUsed only if the database value is not NULL
		if lastUsed.Valid {
			apiKey.LastUsed = &lastUsed.Time
		}

		// Set ExpiresAt only if the database value is not NULL
		if expiresAt.Valid {
			apiKey.ExpiresAt = &expiresAt.Time
		}

		apiKeys = append(apiKeys, apiKey)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating API key rows: %w", err)
	}

	return apiKeys, nil
}

// GetAll retrieves all API keys
func (m *APIKeyModel) GetAll(ctx context.Context) ([]*APIKey, error) {
	query := `
		SELECT id, name, user_id, description, scopes, last_used, created_at, expires_at, is_active
		FROM api_keys
		ORDER BY created_at DESC
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*APIKey
	for rows.Next() {
		var scopesStr string
		var lastUsed sql.NullTime
		var expiresAt sql.NullTime
		apiKey := &APIKey{}
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.Name,
			&apiKey.UserID,
			&apiKey.Description,
			&scopesStr,
			&lastUsed,
			&apiKey.CreatedAt,
			&expiresAt,
			&apiKey.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key row: %w", err)
		}

		// Convert scopes string to array
		if scopesStr != "" {
			apiKey.Scopes = splitScopesString(scopesStr)
		}

		// Set LastUsed only if the database value is not NULL
		if lastUsed.Valid {
			apiKey.LastUsed = &lastUsed.Time
		}

		// Set ExpiresAt only if the database value is not NULL
		if expiresAt.Valid {
			apiKey.ExpiresAt = &expiresAt.Time
		}

		apiKeys = append(apiKeys, apiKey)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating API key rows: %w", err)
	}

	return apiKeys, nil
}

// Deactivate deactivates an API key
func (m *APIKeyModel) Deactivate(ctx context.Context, id string) error {
	query := `
		UPDATE api_keys
		SET is_active = false
		WHERE id = $1
	`

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// Delete deletes an API key
func (m *APIKeyModel) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM api_keys
		WHERE id = $1
	`

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// GenerateAPIKey generates a unique API key
func GenerateAPIKey() string {
	return uuid.New().String() + uuid.New().String()
}

// Helper function to split a comma-separated string of scopes
func splitScopesString(scopesStr string) []string {
	if scopesStr == "" {
		return []string{}
	}

	var scopes []string
	for _, scope := range splitString(scopesStr, ',') {
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

// Helper function to split a string
func splitString(s string, sep rune) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
