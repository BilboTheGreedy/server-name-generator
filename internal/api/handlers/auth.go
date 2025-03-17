package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	custommw "github.com/bilbothegreedy/server-name-generator/internal/api/middleware"
	"github.com/bilbothegreedy/server-name-generator/internal/auth"
	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userModel  *models.UserModel
	jwtManager *auth.JWTManager
	logger     *utils.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userModel *models.UserModel, jwtManager *auth.JWTManager, logger *utils.Logger) *AuthHandler {
	return &AuthHandler{
		userModel:  userModel,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string       `json:"token"`
	User      *models.User `json:"user"`
	ExpiresAt time.Time    `json:"expiresAt"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
	Email    string `json:"email" validate:"required,email"`
}

// Login handles the login request
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode login request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.Validate(req); err != nil {
		h.logger.Error("Invalid login request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Authenticate user
	user, err := h.userModel.AuthenticateUser(r.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Info("Failed login attempt", "username", req.Username, "error", err)
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(user)
	if err != nil {
		h.logger.Error("Failed to generate token", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}

	// Remove password from user object
	user.Password = ""

	// Create response
	expiresAt := time.Now().UTC().Add(24 * time.Hour) // Token duration (match with JWTManager)
	response := LoginResponse{
		Token:     token,
		User:      user,
		ExpiresAt: expiresAt,
	}

	h.logger.Info("User logged in successfully", "username", user.Username)
	utils.RespondWithJSON(w, http.StatusOK, response)
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode register request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.Validate(req); err != nil {
		h.logger.Error("Invalid register request", "error", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if username is already taken
	existingUser, err := h.userModel.GetByUsername(r.Context(), req.Username)
	if err != nil {
		h.logger.Error("Failed to check existing user", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process registration")
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process registration")
		return
	}

	// Create new user
	now := time.Now().UTC()
	user := &models.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		Role:      models.RoleUser, // Default role for new users
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

	h.logger.Info("User registered successfully", "username", user.Username)
	utils.RespondWithJSON(w, http.StatusCreated, user)
}

// GetCurrentUser returns the currently authenticated user
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context
	claims, ok := r.Context().Value(custommw.UserClaimsKey).(*auth.UserClaims)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get user from database
	user, err := h.userModel.GetByID(r.Context(), claims.ID)
	if err != nil {
		h.logger.Error("Failed to get user", "error", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user information")
		return
	}

	if user == nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Remove password from response
	user.Password = ""

	utils.RespondWithJSON(w, http.StatusOK, user)
}
