package models

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Reservation statuses
const (
	StatusReserved  = "reserved"
	StatusCommitted = "committed"
)

// ReservationPayload represents the incoming data for a reservation request
type ReservationPayload struct {
	UnitCode    string `json:"unitCode" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Provider    string `json:"provider" validate:"required"`
	Region      string `json:"region" validate:"required"`
	Environment string `json:"environment" validate:"required"`
	Function    string `json:"function" validate:"required"`
}

// CommitPayload represents the incoming data for a commit request
type CommitPayload struct {
	ReservationID string `json:"reservationId" validate:"required,uuid"`
}

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

// ReservationResponse is the API response for reservation operations
type ReservationResponse struct {
	ReservationID string `json:"reservationId"`
	ServerName    string `json:"serverName"`
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
	query := `
		INSERT INTO reservations (
			id, server_name, unit_code, type, provider, region, environment, function, 
			sequence_num, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		r.ID,
		r.ServerName,
		r.UnitCode,
		r.Type,
		r.Provider,
		r.Region,
		r.Environment,
		r.Function,
		r.SequenceNum,
		r.Status,
		r.CreatedAt,
		r.UpdatedAt,
	)

	return err
}

// GetByID retrieves a reservation by its ID
func (m *ReservationModel) GetByID(ctx context.Context, id string) (*Reservation, error) {
	query := `
		SELECT id, server_name, unit_code, type, provider, region, environment, function,
			   sequence_num, status, created_at, updated_at
		FROM reservations
		WHERE id = $1
	`

	r := &Reservation{}
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&r.ID,
		&r.ServerName,
		&r.UnitCode,
		&r.Type,
		&r.Provider,
		&r.Region,
		&r.Environment,
		&r.Function,
		&r.SequenceNum,
		&r.Status,
		&r.CreatedAt,
		&r.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r, nil
}

// UpdateStatus updates the status of a reservation
func (m *ReservationModel) UpdateStatus(ctx context.Context, tx *sql.Tx, id, status string) error {
	query := `
		UPDATE reservations
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status != $4
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		status,
		time.Now().UTC(),
		id,
		StatusCommitted, // Prevent updating if already committed
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("reservation not found or already committed")
	}

	return nil
}

// IsServerNameUnique checks if a server name is already in use
func (m *ReservationModel) IsServerNameUnique(ctx context.Context, tx *sql.Tx, serverName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM reservations 
			WHERE server_name = $1 AND status = $2
		)
	`

	var exists bool
	err := tx.QueryRowContext(ctx, query, serverName, StatusCommitted).Scan(&exists)
	if err != nil {
		return false, err
	}

	return !exists, nil
}
