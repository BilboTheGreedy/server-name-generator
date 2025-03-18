package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
	"github.com/go-chi/chi/v5"
)

// APIReserve handles the API request to reserve a server name
func (app *Application) APIReserve(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var payload models.ReservationPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.Logger.Error("Failed to decode reserve request body", "error", err)
		app.ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate payload fields
	if err := validateReservationPayload(payload); err != nil {
		app.Logger.Error("Invalid reservation payload", "error", err)
		app.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Reserve the server name
	result, err := app.NameService.ReserveServerName(r.Context(), payload)
	if err != nil {
		app.Logger.Error("Failed to reserve server name", "error", err)

		// Check for specific error cases
		if strings.Contains(err.Error(), "already in use") {
			app.ErrorResponse(w, http.StatusConflict, "Server name is already in use")
			return
		}

		app.ErrorResponse(w, http.StatusInternalServerError, "Failed to reserve server name")
		return
	}

	// Log the successful reservation
	app.Logger.Info("Server name reserved", 
		"reservationId", result.ReservationID,
		"serverName", result.ServerName)

	// Return success response
	app.JSONResponse(w, http.StatusCreated, result)
}

// APICommit handles the API request to commit a reserved server name
func (app *Application) APICommit(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var payload models.CommitPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.Logger.Error("Failed to decode commit request body", "error", err)
		app.ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate reservation ID
	if payload.ReservationID == "" {
		app.ErrorResponse(w, http.StatusBadRequest, "Reservation ID is required")
		return
	}

	// Commit the reservation
	err = app.NameService.CommitReservation(r.Context(), payload.ReservationID)
	if err != nil {
		app.Logger.Error("Failed to commit reservation", "error", err, "reservationId", payload.ReservationID)

		// Check for specific error cases
		if strings.Contains(err.Error(), "already committed") {
			app.ErrorResponse(w, http.StatusConflict, "Reservation is already committed")
			return
		}

		if strings.Contains(err.Error(), "not found") {
			app.ErrorResponse(w, http.StatusNotFound, "Reservation not found")
			return
		}

		app.ErrorResponse(w, http.StatusInternalServerError, "Failed to commit reservation")
		return
	}

	// Log the successful commit
	app.Logger.Info("Reservation committed successfully", "reservationId", payload.ReservationID)

	// Return success response
	app.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Reservation committed successfully",
	})
}

// APIRelease handles the API request to release a committed server name back to reserved status
func (app *Application) APIRelease(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var payload models.ReleasePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.Logger.Error("Failed to decode release request body", "error", err)
		app.ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate reservation ID
	if payload.ReservationID == "" {
		app.ErrorResponse(w, http.StatusBadRequest, "Reservation ID is required")
		return
	}

	// Release the reservation
	err = app.NameService.ReleaseReservation(r.Context(), payload.ReservationID)
	if err != nil {
		app.Logger.Error("Failed to release reservation", "error", err, "reservationId", payload.ReservationID)

		// Check for specific error cases
		if strings.Contains(err.Error(), "not committed") {
			app.ErrorResponse(w, http.StatusBadRequest, "Reservation is not committed")
			return
		}

		if strings.Contains(err.Error(), "not found") {
			app.ErrorResponse(w, http.StatusNotFound, "Reservation not found")
			return
		}

		app.ErrorResponse(w, http.StatusInternalServerError, "Failed to release reservation")
		return
	}

	// Log the successful release
	app.Logger.Info("Reservation released successfully", "reservationId", payload.ReservationID)

	// Return success response
	app.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Reservation released successfully",
	})
}

// APIGetReservations handles the API request to get all reservations
func (app *Application) APIGetReservations(w http.ResponseWriter, r *http.Request) {
	// Get all reservations
	reservations, err := app.NameService.GetAllReservations(r.Context())
	if err != nil {
		app.Logger.Error("Failed to get reservations", "error", err)
		app.ErrorResponse(w, http.StatusInternalServerError, "Failed to get reservations")
		return
	}

	// Return the reservations
	app.JSONResponse(w, http.StatusOK, reservations)
}

// APIDeleteReservation handles the API request to delete a reservation
func (app *Application) APIDeleteReservation(w http.ResponseWriter, r *http.Request) {
	// Get reservation ID from URL
	id := chi.URLParam(r, "id")
	if id == "" {
		app.ErrorResponse(w, http.StatusBadRequest, "Missing reservation ID")
		return
	}

	// Delete the reservation
	err := app.NameService.DeleteReservation(r.Context(), id)
	if err != nil {
		app.Logger.Error("Failed to delete reservation", "error", err, "id", id)

		// Check for specific error cases
		if strings.Contains(err.Error(), "not found") {
			app.ErrorResponse(w, http.StatusNotFound, "Reservation not found")
			return
		}

		if strings.Contains(err.Error(), "cannot delete a committed") {
			app.ErrorResponse(w, http.StatusBadRequest, "Cannot delete a committed reservation. Please release it first.")
			return
		}

		app.ErrorResponse(w, http.StatusInternalServerError, "Failed to delete reservation")
		return
	}

	// Log the successful deletion
	app.Logger.Info("Reservation deleted successfully", "id", id)

	// Return success response
	app.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Reservation deleted successfully",
	})
}

// APIGetStats handles the API request to get usage statistics
func (app *Application) APIGetStats(w http.ResponseWriter, r *http.Request) {
	// Get statistics
	stats, err := app.NameService.GetStats(r.Context())
	if err != nil {
		app.Logger.Error("Failed to get stats", "error", err)
		app.ErrorResponse(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	// Return the statistics
	app.JSONResponse(w, http.StatusOK, stats)
}

// validateReservationPayload validates the reservation payload
func validateReservationPayload(payload models.ReservationPayload) error {
	// Check field lengths
	if len(payload.UnitCode) > 3 {
		return utils.ValidationError("Unit code must be at most 3 characters")
	}
	if len(payload.Type) > 1 {
		return utils.ValidationError("Type must be at most 1 character")
	}
	if len(payload.Provider) > 1 {
		return utils.ValidationError("Provider must be at most 1 character")
	}
	if len(payload.Region) > 4 {
		return utils.ValidationError("Region must be at most 4 characters")
	}
	if len(payload.Environment) > 1 {
		return utils.ValidationError("Environment must be at most 1 character")
	}
	if len(payload.Function) > 2 {
		return utils.ValidationError("Function must be at most 2 characters")
	}

	return nil
}