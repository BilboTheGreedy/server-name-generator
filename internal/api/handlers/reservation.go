package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/services"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// ReservationHandler handles reservation-related HTTP requests
type ReservationHandler struct {
	nameService *services.NameGeneratorService
	logger      *utils.Logger
}

// NewReservationHandler creates a new reservation handler
func NewReservationHandler(nameService *services.NameGeneratorService, logger *utils.Logger) *ReservationHandler {
	return &ReservationHandler{
		nameService: nameService,
		logger:      logger,
	}
}

// Reserve handles POST /reserve requests
func (h *ReservationHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var payload models.ReservationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate payload
	if err := utils.Validate(payload); err != nil {
		h.logger.Error("Invalid reservation payload", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Reserve server name
	result, err := h.nameService.ReserveServerName(r.Context(), payload)
	if err != nil {
		h.logger.Error("Failed to reserve server name", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to reserve server name")
		return
	}

	h.logger.Info("Server name reserved",
		"reservationId", result.ReservationID,
		"serverName", result.ServerName,
	)

	// Respond with reservation details
	utils.RespondWithJSON(w, http.StatusCreated, result)
}
