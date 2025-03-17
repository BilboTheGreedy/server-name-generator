package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/bilbothegreedy/server-name-generator/internal/auth"
	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)
// APIKeyClaimsKey is the key for API key claims in context
const APIKeyClaimsKey ContextKey = "api_key_claims"

// APIKeyClaims represents the claims for an API key
type APIKeyClaims struct {
	KeyID    string
	UserID   string
	Scopes   []string
	IsActive bool
}

// APIKeyAuthenticate middleware verifies API keys and adds claims to request context
func APIKeyAuthenticate(apiKeyModel *models.APIKeyModel, logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for API key in headers
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// Check for API key in query parameters
				apiKey = r.URL.Query().Get("api_key")
				if apiKey == "" {
					// If no API key is found, don't authenticate and continue
					// This allows the route to use other authentication methods or reject unauthenticated requests
					next.ServeHTTP(w, r)
					return
				}
			}

			// Verify API key
			key, err := apiKeyModel.GetByKey(r.Context(), apiKey)
			if err != nil {
				logger.Error("Failed to verify API key", "error", err)
				utils.RespondWithError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}

			if key == nil || !key.IsActive {
				logger.Info("Invalid or inactive API key used")
				utils.RespondWithError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}

			// Create API key claims
			claims := &APIKeyClaims{
				KeyID:    key.ID,
				UserID:   key.UserID,
				Scopes:   key.Scopes,
				IsActive: key.IsActive,
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), APIKeyClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAPIKeyScope middleware requires specific scopes for API keys
func RequireAPIKeyScope(scopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request has API key claims
			claims, ok := r.Context().Value(APIKeyClaimsKey).(*APIKeyClaims)
			if !ok {
				// No API key claims, let the request continue
				// This allows the route to use other authentication methods
				next.ServeHTTP(w, r)
				return
			}

			// Check if API key has required scopes
			if !hasScope(claims.Scopes, scopes) {
				utils.RespondWithError(w, http.StatusForbidden, "API key missing required scope")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper function to check if the API key has the required scopes
func hasScope(keyScopes []string, requiredScopes []string) bool {
	// If no scopes are required, return true
	if len(requiredScopes) == 0 {
		return true
	}

	// Check if the API key has any of the required scopes
	for _, required := range requiredScopes {
		for _, scope := range keyScopes {
			// Support wildcard scopes (e.g., "admin:*" matches "admin:read", "admin:write", etc.)
			if strings.HasSuffix(scope, "*") {
				prefix := strings.TrimSuffix(scope, "*")
				if strings.HasPrefix(required, prefix) {
					return true
				}
			} else if scope == required {
				return true
			}
		}
	}

	return false
}

// GetAPIKeyClaims retrieves API key claims from request context
func GetAPIKeyClaims(r *http.Request) (*APIKeyClaims, bool) {
	claims, ok := r.Context().Value(APIKeyClaimsKey).(*APIKeyClaims)
	return claims, ok
}

// CombinedAuth middleware tries both JWT and API key authentication
func CombinedAuth(jwtManager *auth.JWTManager, apiKeyModel *models.APIKeyModel, logger *utils.Logger) func(http.Handler) http.Handler {
	jwtAuth := Authenticate(jwtManager, logger)
	apiKeyAuth := APIKeyAuthenticate(apiKeyModel, logger)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new response writer that captures status code
			wrappedWriter := utils.NewResponseWriter(w)

			// Try JWT authentication first
			var authenticated bool
			var tempHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// If we reached this handler, JWT auth was successful
				authenticated = true
				next.ServeHTTP(w, r)
			})

			// Apply JWT authentication
			jwtAuth(tempHandler).ServeHTTP(wrappedWriter, r)

			// If JWT authentication successful or an error response was sent, return
			if authenticated || wrappedWriter.StatusCode != 0 {
				return
			}

			// Reset writer status
			wrappedWriter.StatusCode = 0

			// If JWT authentication failed, try API key authentication
			apiKeyAuth(next).ServeHTTP(w, r)
		})
	}
}
