// DeleteReservation deletes a reservation by ID (now works for both reserved and committed)
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

	// Delete the reservation regardless of status
	if err := s.reservationModel.Delete(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete reservation: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}