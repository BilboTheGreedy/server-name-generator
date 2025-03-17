package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/bilbothegreedy/server-name-generator/internal/errors"
	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
	"github.com/google/uuid"
)

// Character limits for each field
const (
	UnitCodeMaxLen    = 3
	TypeMaxLen        = 1
	ProviderMaxLen    = 1
	RegionMaxLen      = 4
	EnvironmentMaxLen = 1
	FunctionMaxLen    = 2
	SequenceMaxLen    = 3
)

// Stats represents dashboard statistics
type Stats struct {
	TotalReservations  int                   `json:"totalReservations"`
	CommittedCount     int                   `json:"committedCount"`
	ReservedCount      int                   `json:"reservedCount"`
	RecentReservations []*models.Reservation `json:"recentReservations"`
	TopEnvironments    []EnvStat             `json:"topEnvironments"`
	TopRegions         []RegionStat          `json:"topRegions"`
	DailyActivity      []DailyStat           `json:"dailyActivity"`
}

// EnvStat represents environment usage statistics
type EnvStat struct {
	Environment string `json:"environment"`
	Count       int    `json:"count"`
}

// RegionStat represents region usage statistics
type RegionStat struct {
	Region string `json:"region"`
	Count  int    `json:"count"`
}

// DailyStat represents daily activity statistics
type DailyStat struct {
	Date      string `json:"date"`
	Reserved  int    `json:"reserved"`
	Committed int    `json:"committed"`
}

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

// NormalizeField truncates or pads a field to its required length
func (s *NameGeneratorService) NormalizeField(value string, maxLen int, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	// Convert to uppercase
	value = strings.ToUpper(value)

	// Truncate if longer than maxLen
	if len(value) > maxLen {
		return value[:maxLen]
	}

	return value
}

// GenerateServerName creates a server name from the provided parameters and sequence number
func (s *NameGeneratorService) GenerateServerName(params models.ReservationPayload, sequenceNum int) string {
	// Normalize each field to its max length
	unitCode := s.NormalizeField(params.UnitCode, UnitCodeMaxLen, "SRV")
	hostType := s.NormalizeField(params.Type, TypeMaxLen, "V")          // Default V for VM
	provider := s.NormalizeField(params.Provider, ProviderMaxLen, "X")  // Default X for Mixed
	region := s.NormalizeField(params.Region, RegionMaxLen, "GLBL")     // Default GLBL for Global
	env := s.NormalizeField(params.Environment, EnvironmentMaxLen, "P") // Default P for Production
	function := s.NormalizeField(params.Function, FunctionMaxLen, "SV") // Default SV for Server

	// Format the sequence number with leading zeros
	sequenceStr := fmt.Sprintf("%0*d", SequenceMaxLen, sequenceNum)

	// Ensure sequence doesn't exceed maximum length
	if len(sequenceStr) > SequenceMaxLen {
		sequenceStr = sequenceStr[len(sequenceStr)-SequenceMaxLen:]
	}

	// Combine all parts with no separators to form a fixed-width name
	return fmt.Sprintf("%s%s%s%s%s%s%s",
		unitCode,
		hostType,
		provider,
		region,
		env,
		function,
		sequenceStr,
	)
}

// GetNameBasePattern creates a pattern for finding similar server names
func (s *NameGeneratorService) GetNameBasePattern(params models.ReservationPayload) string {
	// Normalize each field to its max length
	unitCode := s.NormalizeField(params.UnitCode, UnitCodeMaxLen, "SRV")
	hostType := s.NormalizeField(params.Type, TypeMaxLen, "V")
	provider := s.NormalizeField(params.Provider, ProviderMaxLen, "X")
	region := s.NormalizeField(params.Region, RegionMaxLen, "GLBL")
	env := s.NormalizeField(params.Environment, EnvironmentMaxLen, "P")
	function := s.NormalizeField(params.Function, FunctionMaxLen, "SV")

	// Create pattern without sequence number
	return fmt.Sprintf("%s%s%s%s%s%s",
		unitCode,
		hostType,
		provider,
		region,
		env,
		function,
	)
}

func (s *NameGeneratorService) ReserveServerName(ctx context.Context, params models.ReservationPayload) (*models.ReservationResponse, error) {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, errors.NewDatabaseError("Failed to start transaction", err)
	}
	defer tx.Rollback() // Will be ignored if transaction is committed

	// Get the base pattern for the name
	nameBasePattern := s.GetNameBasePattern(params)

	// Find the latest sequence number for similar names
	latestSequence, err := s.reservationModel.FindLatestSequenceNumber(ctx, tx, nameBasePattern)
	if err != nil {
		return nil, errors.NewDatabaseError("Failed to find latest sequence number", err)
	}

	// Increment sequence number
	sequenceNum := latestSequence + 1

	// Generate server name
	serverName := s.GenerateServerName(params, sequenceNum)

	// Check if server name is unique (committed reservations)
	isUnique, err := s.reservationModel.IsServerNameUnique(ctx, tx, serverName)
	if err != nil {
		return nil, errors.NewDatabaseError("Failed to check server name uniqueness", err)
	}

	if !isUnique {
		return nil, errors.NewConflictError(fmt.Sprintf("Server name %s is already in use", serverName))
	}

	// Create reservation
	now := time.Now().UTC()
	reservationID := uuid.New().String()

	// Normalize fields for storage
	reservation := &models.Reservation{
		ID:          reservationID,
		ServerName:  serverName,
		UnitCode:    s.NormalizeField(params.UnitCode, UnitCodeMaxLen, "SRV"),
		Type:        s.NormalizeField(params.Type, TypeMaxLen, "V"),
		Provider:    s.NormalizeField(params.Provider, ProviderMaxLen, "X"),
		Region:      s.NormalizeField(params.Region, RegionMaxLen, "GLBL"),
		Environment: s.NormalizeField(params.Environment, EnvironmentMaxLen, "P"),
		Function:    s.NormalizeField(params.Function, FunctionMaxLen, "SV"),
		SequenceNum: sequenceNum,
		Status:      models.StatusReserved,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.reservationModel.Create(ctx, tx, reservation); err != nil {
		return nil, errors.NewDatabaseError("Failed to create reservation", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, errors.NewDatabaseError("Failed to commit transaction", err)
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

// GetAllReservations retrieves all reservations
func (s *NameGeneratorService) GetAllReservations(ctx context.Context) ([]*models.Reservation, error) {
	query := `
		SELECT id, server_name, unit_code, type, provider, region, environment, function,
			   sequence_num, status, created_at, updated_at
		FROM reservations
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*models.Reservation
	for rows.Next() {
		r := &models.Reservation{}
		err := rows.Scan(
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
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %w", err)
	}

	return reservations, nil
}

// DeleteReservation deletes a reservation by ID (only if not committed)
func (s *NameGeneratorService) DeleteReservation(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// First check if the reservation exists
	reservation, err := s.reservationModel.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	if reservation == nil {
		return fmt.Errorf("reservation with ID %s not found", id)
	}

	// Check if the reservation is committed
	if reservation.Status == "committed" {
		return fmt.Errorf("cannot delete a committed reservation")
	}

	// Delete the reservation
	if err := s.reservationModel.Delete(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete reservation: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Reservation deleted successfully", "id", id, "serverName", reservation.ServerName)
	return nil
}

// GetStats retrieves statistics for the admin dashboard
func (s *NameGeneratorService) GetStats(ctx context.Context) (*Stats, error) {
	stats := &Stats{}

	// Get counts of reservations by status
	countQuery := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'committed' THEN 1 ELSE 0 END) as committed,
			SUM(CASE WHEN status = 'reserved' THEN 1 ELSE 0 END) as reserved
		FROM reservations
	`
	err := s.db.QueryRowContext(ctx, countQuery).Scan(
		&stats.TotalReservations,
		&stats.CommittedCount,
		&stats.ReservedCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation counts: %w", err)
	}

	// Get recent reservations (limited to 10)
	recentQuery := `
		SELECT id, server_name, unit_code, type, provider, region, environment, function,
			   sequence_num, status, created_at, updated_at
		FROM reservations
		ORDER BY created_at DESC
		LIMIT 10
	`
	rows, err := s.db.QueryContext(ctx, recentQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent reservations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		r := &models.Reservation{}
		err := rows.Scan(
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
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		stats.RecentReservations = append(stats.RecentReservations, r)
	}

	// Get top environments
	envQuery := `
		SELECT environment, COUNT(*) as count
		FROM reservations
		GROUP BY environment
		ORDER BY count DESC
		LIMIT 5
	`
	envRows, err := s.db.QueryContext(ctx, envQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query top environments: %w", err)
	}
	defer envRows.Close()

	for envRows.Next() {
		var env EnvStat
		err := envRows.Scan(&env.Environment, &env.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan environment stat: %w", err)
		}
		stats.TopEnvironments = append(stats.TopEnvironments, env)
	}

	// Get top regions
	regionQuery := `
		SELECT region, COUNT(*) as count
		FROM reservations
		GROUP BY region
		ORDER BY count DESC
		LIMIT 5
	`
	regionRows, err := s.db.QueryContext(ctx, regionQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query top regions: %w", err)
	}
	defer regionRows.Close()

	for regionRows.Next() {
		var region RegionStat
		err := regionRows.Scan(&region.Region, &region.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan region stat: %w", err)
		}
		stats.TopRegions = append(stats.TopRegions, region)
	}

	// Get daily activity for the last 7 days
	dailyQuery := `
		SELECT 
			TO_CHAR(DATE_TRUNC('day', created_at), 'YYYY-MM-DD') as date,
			SUM(CASE WHEN status = 'reserved' THEN 1 ELSE 0 END) as reserved,
			SUM(CASE WHEN status = 'committed' THEN 1 ELSE 0 END) as committed
		FROM reservations
		WHERE created_at >= NOW() - INTERVAL '7 days'
		GROUP BY DATE_TRUNC('day', created_at)
		ORDER BY date
	`
	dailyRows, err := s.db.QueryContext(ctx, dailyQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily activity: %w", err)
	}
	defer dailyRows.Close()

	for dailyRows.Next() {
		var day DailyStat
		err := dailyRows.Scan(&day.Date, &day.Reserved, &day.Committed)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily stat: %w", err)
		}
		stats.DailyActivity = append(stats.DailyActivity, day)
	}

	return stats, nil
}

// ReleaseReservation changes a reservation status from committed to reserved
func (s *NameGeneratorService) ReleaseReservation(ctx context.Context, id string) error {
	// First check if the reservation exists and is committed
	reservation, err := s.reservationModel.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	if reservation == nil {
		return fmt.Errorf("reservation with ID %s not found", id)
	}

	if reservation.Status != "committed" {
		return fmt.Errorf("reservation is not committed")
	}

	// Start a transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Release the reservation
	if err := s.reservationModel.Release(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to release reservation: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Reservation released successfully",
		"id", id,
		"serverName", reservation.ServerName,
	)

	return nil
}
