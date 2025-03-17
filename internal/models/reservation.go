// Delete deletes a reservation by ID (works for any status)
func (m *ReservationModel) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	query := `
		DELETE FROM reservations
		WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete reservation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reservation not found")
	}

	return nil
}