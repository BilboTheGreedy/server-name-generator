package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/services"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// CommitHandler handles commit-related HTTP requests
type CommitHandler struct {
	nameService *services.NameGeneratorService
	logger      *utils.Logger
}

// NewCommitHandler creates a new commit handler
func NewCommitHandler(nameService *services.NameGeneratorService, logger *utils.Logger) *CommitHandler {
	return &CommitHandler{
		nameService: nameService,
		logger:      logger,
	}
}

// Commit handles POST /commit requests
func (h *CommitHandler) Commit(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var payload models.CommitPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logger.Error("Failed to decode commit request body", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate payload
	if err := utils.Validate(payload); err != nil {
		h.logger.Error("Invalid commit payload", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Commit the reservation
	if err := h.nameService.CommitReservation(r.Context(), payload.ReservationID); err != nil {
		h.logger.Error("Failed to commit reservation", "error", err, "reservationId", payload.ReservationID)

		// Check for specific error cases and return appropriate status codes
		if err.Error() == "reservation is already committed" {
			utils.RespondWithError(w, http.StatusConflict, "Reservation is already committed")
			return
		}

		if err.Error() == "reservation with ID "+payload.ReservationID+" not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Reservation not found")
			return
		}

		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to commit reservation")
		return
	}

	h.logger.Info("Reservation committed successfully", "reservationId", payload.ReservationID)

	// Respond with success message
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Reservation committed successfully",
	})
}
