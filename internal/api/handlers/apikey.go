package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	custommw "github.com/bilbothegreedy/server-name-generator/internal/api/middleware"
	"github.com/bilbothegreedy/server-name-generator/internal/auth"
	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// APIKeyHandler handles API key-related HTTP requests
type APIKeyHandler struct {
	apiKeyModel *models.APIKeyModel
	userModel   *models.UserModel
	logger      *utils.Logger
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyModel *models.APIKeyModel, userModel *models.UserModel, logger *utils.Logger) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyModel: apiKeyModel,
		userModel:   userModel,
		logger:      logger,
	}
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Scopes      []string `json:"scopes"`
	ExpiresIn   int      `json:"expiresIn"` // in days, 0 means no expiration
}

// GetAll returns all API keys (admin only)
func (h *APIKeyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	keys, err := h.apiKeyModel.GetAll(r.Context())
	if err != nil {
		h.logger.Error("Failed to get API keys", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get API keys")
		return
	}

	// Get user details for each key
	type KeyWithUser struct {
		*models.APIKey
		Username string `json:"username"`
	}

	keysWithUsers := make([]KeyWithUser, 0, len(keys))
	for _, key := range keys {
		user, err := h.userModel.GetByID(r.Context(), key.UserID)
		if err != nil {
			h.logger.Error("Failed to get user for API key", "error", err, "userId", key.UserID)
			continue
		}

		username := "Unknown"
		if user != nil {
			username = user.Username
		}

		keysWithUsers = append(keysWithUsers, KeyWithUser{
			APIKey:   key,
			Username: username,
		})
	}

	utils.RespondWithJSON(w, http.StatusOK, keysWithUsers)
}

// GetAllForUser returns all API keys for the current user
func (h *APIKeyHandler) GetAllForUser(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context
	claims, ok := r.Context().Value(custommw.UserClaimsKey).(*auth.UserClaims)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User claims not found")
		return
	}

	keys, err := h.apiKeyModel.GetByUserID(r.Context(), claims.ID)
	if err != nil {
		h.logger.Error("Failed to get API keys", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get API keys")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, keys)
}

// Create creates a new API key
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode create API key request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.Validate(req); err != nil {
		h.logger.Error("Invalid create API key request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get user claims from context
	claims, ok := r.Context().Value(custommw.UserClaimsKey).(*auth.UserClaims)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User claims not found")
		return
	}

	// Create API key
	now := time.Now().UTC()
	expiresAt := time.Time{}
	if req.ExpiresIn > 0 {
		expiresAt = now.AddDate(0, 0, req.ExpiresIn)
	}

	apiKey := &models.APIKey{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Key:         models.GenerateAPIKey(),
		UserID:      claims.ID,
		Description: req.Description,
		Scopes:      req.Scopes,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		IsActive:    true,
	}

	// Save API key to database
	if err := h.apiKeyModel.Create(r.Context(), nil, apiKey); err != nil {
		h.logger.Error("Failed to create API key", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	h.logger.Info("API key created successfully", "keyId", apiKey.ID, "userId", claims.ID)
	utils.RespondWithJSON(w, http.StatusCreated, apiKey)
}

// Revoke revokes an API key
func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	// Get key ID from URL parameter
	keyID := chi.URLParam(r, "id")
	if keyID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing API key ID")
		return
	}

	// Get user claims from context
	claims, ok := r.Context().Value(custommw.UserClaimsKey).(*auth.UserClaims)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "User claims not found")
		return
	}

	// Get API key to check ownership
	key, err := h.apiKeyModel.GetByID(r.Context(), keyID)
	if err != nil {
		h.logger.Error("Failed to get API key", "error", err, "keyId", keyID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get API key")
		return
	}

	if key == nil {
		utils.RespondWithError(w, http.StatusNotFound, "API key not found")
		return
	}

	// Check if user owns the key or is an admin
	if key.UserID != claims.ID && claims.Role != models.RoleAdmin {
		utils.RespondWithError(w, http.StatusForbidden, "You don't have permission to revoke this API key")
		return
	}

	// Revoke API key
	if err := h.apiKeyModel.Deactivate(r.Context(), keyID); err != nil {
		h.logger.Error("Failed to revoke API key", "error", err, "keyId", keyID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to revoke API key")
		return
	}

	h.logger.Info("API key revoked successfully", "keyId", keyID, "userId", claims.ID)
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "API key revoked successfully",
	})
}

// Delete deletes an API key (admin only)
func (h *APIKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Get key ID from URL parameter
	keyID := chi.URLParam(r, "id")
	if keyID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing API key ID")
		return
	}

	// Delete API key
	if err := h.apiKeyModel.Delete(r.Context(), keyID); err != nil {
		h.logger.Error("Failed to delete API key", "error", err, "keyId", keyID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete API key")
		return
	}

	h.logger.Info("API key deleted successfully", "keyId", keyID)
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "API key deleted successfully",
	})
}
