package db

import (
	"context"
	"database/sql"
	"fmt"
)

// ExecuteInTransaction runs the given function in a database transaction
func ExecuteInTransaction(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	// Start transaction
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer a rollback in case anything fails
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // Re-throw panic after rolling back
		} else if err != nil {
			_ = tx.Rollback() // err is non-nil, rollback
		}
	}()

	// Execute the function in the transaction
	err = fn(tx)
	if err != nil {
		return err // Rollback happens in the deferred function
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
