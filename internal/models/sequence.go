package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// SequenceKey defines a unique key to identify a naming sequence
type SequenceKey struct {
	UnitCode    string
	Type        string
	Provider    string
	Region      string
	Environment string
	Function    string
}

// SequenceModel handles database operations for naming sequences
type SequenceModel struct {
	DB *sql.DB
}

// NewSequenceModel creates a new sequence model
func NewSequenceModel(db *sql.DB) *SequenceModel {
	return &SequenceModel{DB: db}
}

// GetNextSequenceNumber retrieves and increments the sequence number for a given key
func (m *SequenceModel) GetNextSequenceNumber(ctx context.Context, tx *sql.Tx, key SequenceKey) (int, error) {
	query := `
		INSERT INTO sequences (
			unit_code, type, provider, region, environment, function, current_value
		) VALUES (
			$1, $2, $3, $4, $5, $6, 1
		) 
		ON CONFLICT (unit_code, type, provider, region, environment, function) 
		DO UPDATE SET current_value = sequences.current_value + 1
		RETURNING current_value
	`

	var sequenceNum int
	err := tx.QueryRowContext(
		ctx,
		query,
		key.UnitCode,
		key.Type,
		key.Provider,
		key.Region,
		key.Environment,
		key.Function,
	).Scan(&sequenceNum)

	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence number: %w", err)
	}

	return sequenceNum, nil
}

// GetCurrentSequenceNumber retrieves the current sequence number without incrementing
func (m *SequenceModel) GetCurrentSequenceNumber(ctx context.Context, key SequenceKey) (int, error) {
	query := `
		SELECT current_value 
		FROM sequences 
		WHERE unit_code = $1 
		  AND type = $2 
		  AND provider = $3 
		  AND region = $4 
		  AND environment = $5 
		  AND function = $6
	`

	var sequenceNum int
	err := m.DB.QueryRowContext(
		ctx,
		query,
		key.UnitCode,
		key.Type,
		key.Provider,
		key.Region,
		key.Environment,
		key.Function,
	).Scan(&sequenceNum)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil // No sequence exists yet
		}
		return 0, err
	}

	return sequenceNum, nil
}
