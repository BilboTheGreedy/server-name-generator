package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/bilbothegreedy/server-name-generator/internal/auth"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// Key for storing user claims in context
type ContextKey string

const (
	// UserClaimsKey is the key for user claims in context
	UserClaimsKey ContextKey = "user_claims"
)

// Authenticate middleware verifies JWT tokens and adds user claims to request context
func Authenticate(jwtManager *auth.JWTManager, logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.RespondWithError(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			// Check format: Bearer <token>
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.RespondWithError(w, http.StatusUnauthorized, "Invalid authorization format")
				return
			}

			tokenString := parts[1]

			// Verify token
			claims, err := jwtManager.VerifyToken(tokenString)
			if err != nil {
				logger.Error("Failed to verify JWT token", "error", err)
				utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware requires a specific role to access a route
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user claims from context
			claims, ok := r.Context().Value(UserClaimsKey).(*auth.UserClaims)
			if !ok {
				utils.RespondWithError(w, http.StatusUnauthorized, "User claims not found")
				return
			}

			// Check if user has one of the required roles
			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				utils.RespondWithError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserClaims retrieves user claims from request context
func GetUserClaims(r *http.Request) (*auth.UserClaims, bool) {
	claims, ok := r.Context().Value(UserClaimsKey).(*auth.UserClaims)
	return claims, ok
}

// OptionalAuth middleware tries to authenticate the user but allows requests to continue if authentication fails
func OptionalAuth(jwtManager *auth.JWTManager, logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No auth header, continue without user claims
				next.ServeHTTP(w, r)
				return
			}

			// Check format: Bearer <token>
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				// Invalid format, continue without user claims
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]

			// Try to verify token
			claims, err := jwtManager.VerifyToken(tokenString)
			if err != nil {
				// Invalid token, continue without user claims
				logger.Debug("Failed to verify optional JWT token", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
