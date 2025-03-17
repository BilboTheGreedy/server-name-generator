package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/services"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// ReleaseHandler handles release-related HTTP requests
type ReleaseHandler struct {
	nameService *services.NameGeneratorService
	logger      *utils.Logger
}

// NewReleaseHandler creates a new release handler
func NewReleaseHandler(nameService *services.NameGeneratorService, logger *utils.Logger) *ReleaseHandler {
	return &ReleaseHandler{
		nameService: nameService,
		logger:      logger,
	}
}

// Release handles POST /release requests
func (h *ReleaseHandler) Release(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var payload models.ReleasePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logger.Error("Failed to decode release request body", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate payload
	if err := utils.Validate(payload); err != nil {
		h.logger.Error("Invalid release payload", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Release the reservation
	if err := h.nameService.ReleaseReservation(r.Context(), payload.ReservationID); err != nil {
		h.logger.Error("Failed to release reservation", "error", err, "reservationId", payload.ReservationID)

		// Check for specific error cases and return appropriate status codes
		if err.Error() == "reservation is not committed" {
			utils.RespondWithError(w, http.StatusBadRequest, "Reservation is not committed")
			return
		}

		if err.Error() == "reservation with ID "+payload.ReservationID+" not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Reservation not found")
			return
		}

		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to release reservation")
		return
	}

	h.logger.Info("Reservation released successfully", "reservationId", payload.ReservationID)

	// Respond with success message
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Reservation released successfully",
	})
}
