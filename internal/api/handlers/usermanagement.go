package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserManagementHandler handles user management HTTP requests
type UserManagementHandler struct {
	userModel *models.UserModel
	logger    *utils.Logger
}

// NewUserManagementHandler creates a new user management handler
func NewUserManagementHandler(userModel *models.UserModel, logger *utils.Logger) *UserManagementHandler {
	return &UserManagementHandler{
		userModel: userModel,
		logger:    logger,
	}
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
	Email    string `json:"email" validate:"required,email"`
	Role     string `json:"role" validate:"required,oneof=admin user"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=admin user"`
}

// ChangePasswordRequest represents the request to change a user's password
type ChangePasswordRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}

// GetAllUsers returns all users
func (h *UserManagementHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userModel.GetAll(r.Context())
	if err != nil {
		h.logger.Error("Failed to get users", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, users)
}

// GetUser returns a user by ID
func (h *UserManagementHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL parameter
	userID := chi.URLParam(r, "id")
	if userID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	user, err := h.userModel.GetByID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user", "error", err, "userId", userID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if user == nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// CreateUser creates a new user
func (h *UserManagementHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode create user request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.Validate(req); err != nil {
		h.logger.Error("Invalid create user request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if username is already taken
	existingUser, err := h.userModel.GetByUsername(r.Context(), req.Username)
	if err != nil {
		h.logger.Error("Failed to check existing user", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process user creation")
		return
	}

	if existingUser != nil {
		utils.RespondWithError(w, http.StatusConflict, "Username already taken")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Failed to hash password", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process user creation")
		return
	}

	// Create new user
	now := time.Now().UTC()
	user := &models.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		Role:      req.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save user to database
	if err := h.userModel.Create(r.Context(), nil, user); err != nil {
		h.logger.Error("Failed to create user", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Remove password from response
	user.Password = ""

	h.logger.Info("User created successfully", "username", user.Username, "userId", user.ID, "role", user.Role)
	utils.RespondWithJSON(w, http.StatusCreated, user)
}

// UpdateUser updates a user
func (h *UserManagementHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL parameter
	userID := chi.URLParam(r, "id")
	if userID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	// Decode request body
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode update user request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.Validate(req); err != nil {
		h.logger.Error("Invalid update user request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get user from database
	user, err := h.userModel.GetByID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user", "error", err, "userId", userID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	if user == nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Update user fields if provided
	updated := false
	if req.Username != "" && req.Username != user.Username {
		// Check if new username is available
		existingUser, err := h.userModel.GetByUsername(r.Context(), req.Username)
		if err != nil {
			h.logger.Error("Failed to check existing user", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}

		if existingUser != nil {
			utils.RespondWithError(w, http.StatusConflict, "Username already taken")
			return
		}

		user.Username = req.Username
		updated = true
	}

	if req.Email != "" && req.Email != user.Email {
		user.Email = req.Email
		updated = true
	}

	if req.Role != "" && req.Role != user.Role {
		user.Role = req.Role
		updated = true
	}

	if updated {
		// Update user in database
		user.UpdatedAt = time.Now().UTC()
		if err := h.userModel.Update(r.Context(), user); err != nil {
			h.logger.Error("Failed to update user", "error", err, "userId", userID)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}

		h.logger.Info("User updated successfully", "userId", userID)
	}

	// Remove password from response
	user.Password = ""

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// ChangeUserPassword changes a user's password
func (h *UserManagementHandler) ChangeUserPassword(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL parameter
	userID := chi.URLParam(r, "id")
	if userID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	// Decode request body
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode change password request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.Validate(req); err != nil {
		h.logger.Error("Invalid change password request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user exists
	user, err := h.userModel.GetByID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user", "error", err, "userId", userID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to change password")
		return
	}

	if user == nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Update password
	if err := h.userModel.UpdatePassword(r.Context(), userID, req.Password); err != nil {
		h.logger.Error("Failed to update password", "error", err, "userId", userID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to change password")
		return
	}

	h.logger.Info("Password changed successfully", "userId", userID)
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// DeleteUser deletes a user
func (h *UserManagementHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL parameter
	userID := chi.URLParam(r, "id")
	if userID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	// Delete user
	if err := h.userModel.Delete(r.Context(), userID); err != nil {
		h.logger.Error("Failed to delete user", "error", err, "userId", userID)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	h.logger.Info("User deleted successfully", "userId", userID)
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}
