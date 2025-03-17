package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
	"github.com/google/uuid"
)

// NameGeneratorService handles the business logic for server name generation
type NameGeneratorService struct {
	db               *sql.DB
	sequenceModel    *models.SequenceModel
	reservationModel *models.ReservationModel
	logger           *utils.Logger
}

// NewNameGeneratorService creates a new name generator service
func NewNameGeneratorService(
	db *sql.DB,
	sequenceModel *models.SequenceModel,
	reservationModel *models.ReservationModel,
	logger *utils.Logger,
) *NameGeneratorService {
	return &NameGeneratorService{
		db:               db,
		sequenceModel:    sequenceModel,
		reservationModel: reservationModel,
		logger:           logger,
	}
}

// GenerateServerName creates a server name from the provided parameters and sequence number
func (s *NameGeneratorService) GenerateServerName(params models.ReservationPayload, sequenceNum int) string {
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s-%s-%04d",
		params.UnitCode,
		params.Type,
		params.Provider,
		params.Region,
		params.Environment,
		params.Function,
		sequenceNum,
	)
}

// ReserveServerName reserves a new server name
func (s *NameGeneratorService) ReserveServerName(ctx context.Context, params models.ReservationPayload) (*models.ReservationResponse, error) {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback() // Will be ignored if transaction is committed

	// Get next sequence number
	sequenceKey := models.SequenceKey{
		UnitCode:    params.UnitCode,
		Type:        params.Type,
		Provider:    params.Provider,
		Region:      params.Region,
		Environment: params.Environment,
		Function:    params.Function,
	}

	sequenceNum, err := s.sequenceModel.GetNextSequenceNumber(ctx, tx, sequenceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get sequence number: %w", err)
	}

	// Generate server name
	serverName := s.GenerateServerName(params, sequenceNum)

	// Check if server name is unique (committed reservations)
	isUnique, err := s.reservationModel.IsServerNameUnique(ctx, tx, serverName)
	if err != nil {
		return nil, fmt.Errorf("failed to check server name uniqueness: %w", err)
	}

	if !isUnique {
		return nil, fmt.Errorf("server name %s is already in use", serverName)
	}

	// Create reservation
	now := time.Now().UTC()
	reservationID := uuid.New().String()

	reservation := &models.Reservation{
		ID:          reservationID,
		ServerName:  serverName,
		UnitCode:    params.UnitCode,
		Type:        params.Type,
		Provider:    params.Provider,
		Region:      params.Region,
		Environment: params.Environment,
		Function:    params.Function,
		SequenceNum: sequenceNum,
		Status:      models.StatusReserved,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.reservationModel.Create(ctx, tx, reservation); err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &models.ReservationResponse{
		ReservationID: reservationID,
		ServerName:    serverName,
	}, nil
}

// CommitReservation commits a server name reservation
func (s *NameGeneratorService) CommitReservation(ctx context.Context, reservationID string) error {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback() // Will be ignored if transaction is committed

	// Check if reservation exists
	reservation, err := s.reservationModel.GetByID(ctx, reservationID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	if reservation == nil {
		return fmt.Errorf("reservation with ID %s not found", reservationID)
	}

	if reservation.Status == models.StatusCommitted {
		return fmt.Errorf("reservation is already committed")
	}

	// Update reservation status to committed
	if err := s.reservationModel.UpdateStatus(ctx, tx, reservationID, models.StatusCommitted); err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
