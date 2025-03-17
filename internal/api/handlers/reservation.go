package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bilbothegreedy/server-name-generator/internal/errors"
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
	ctx := r.Context()

	// Decode request body
	var payload models.ReservationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logger.LogError(ctx, err, "Failed to decode request body")
		appErr := errors.NewBadRequestError("Invalid request payload", err)
		utils.RespondWithAppError(w, ctx, appErr)
		return
	}

	// Validate payload
	if err := utils.Validate(payload); err != nil {
		h.logger.LogError(ctx, err, "Invalid reservation payload")
		appErr := errors.NewValidationError(err.Error(), err)
		utils.RespondWithAppError(w, ctx, appErr)
		return
	}

	// Reserve server name
	result, err := h.nameService.ReserveServerName(ctx, payload)
	if err != nil {
		h.logger.LogError(ctx, err, "Failed to reserve server name")

		// Convert error to appropriate type
		var appErr *errors.AppError
		if strings.Contains(err.Error(), "already in use") {
			appErr = errors.NewConflictError("Server name is already in use")
		} else {
			appErr = errors.NewInternalError("Failed to reserve server name", err)
		}

		utils.RespondWithAppError(w, ctx, appErr)
		return
	}

	h.logger.Info("Server name reserved",
		"reservationId", result.ReservationID,
		"serverName", result.ServerName,
	)

	// Respond with reservation details
	utils.RespondWithJSON(w, http.StatusCreated, result)
}
