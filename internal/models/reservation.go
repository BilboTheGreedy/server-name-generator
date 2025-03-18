package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Reservation statuses
const (
	StatusReserved  = "reserved"
	StatusCommitted = "committed"
)

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

// ReservationPayload represents the request payload for reserving a server name
type ReservationPayload struct {
	UnitCode    string `json:"unitCode,omitempty"`
	Type        string `json:"type,omitempty"`
	Provider    string `json:"provider,omitempty"`
	Region      string `json:"region,omitempty"`
	Environment string `json:"environment,omitempty"`
	Function    string `json:"function,omitempty"`
}

// ReservationResponse is the API response for reservation operations
type ReservationResponse struct {
	ReservationID string `json:"reservationId"`
	ServerName    string `json:"serverName"`
}

// CommitPayload represents the request payload for committing a reservation
type CommitPayload struct {
	ReservationID string `json:"reservationId"`
}

// ReleasePayload represents the request payload for releasing a reservation
type ReleasePayload struct {
	ReservationID string `json:"reservationId"`
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
	// Implementation details
	return nil
}

// GetByID retrieves a reservation by its ID
func (m *ReservationModel) GetByID(ctx context.Context, id string) (*Reservation, error) {
	// Implementation details
	return nil, nil
}

// IsServerNameUnique checks if a server name is already in use
func (m *ReservationModel) IsServerNameUnique(ctx context.Context, tx *sql.Tx, serverName string) (bool, error) {
	// Implementation details
	return true, nil
}

// FindLatestSequenceNumber finds the latest sequence number for a similar name pattern
func (m *ReservationModel) FindLatestSequenceNumber(ctx context.Context, tx *sql.Tx, pattern string) (int, error) {
	// Implementation details
	return 0, nil
}

// UpdateStatus updates the status of a reservation
func (m *ReservationModel) UpdateStatus(ctx context.Context, tx *sql.Tx, id, status string) error {
	// Implementation details
	return nil
}

// Delete deletes a reservation by ID (works for any status)
func (m *ReservationModel) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	// Implementation details
	return nil
}

// Release changes a committed reservation back to reserved
func (m *ReservationModel) Release(ctx context.Context, tx *sql.Tx, id string) error {
	// Implementation details
	return nil
}
