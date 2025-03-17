package models

import (
	"time"
)

// RegisterPayload represents the request payload for user registration
type RegisterPayload struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

// LoginPayload represents the request payload for user login
type LoginPayload struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// TokenResponse represents the response for authentication operations
type TokenResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	User         UserInfo  `json:"user"`
}

// UserInfo represents the user information returned with tokens
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// RefreshTokenPayload represents the request payload for refreshing tokens
type RefreshTokenPayload struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// ChangePasswordPayload represents the request payload for changing password
type ChangePasswordPayload struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,max=100"`
}

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"tokenType"` // access or refresh
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}
