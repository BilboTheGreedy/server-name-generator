// internal/api/middleware/request_id.go
package middleware

import (
	"context"
	"net/http"

	"github.com/bilbothegreedy/server-name-generator/internal/utils"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from header or generate a new one
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Set request ID in header
			w.Header().Set("X-Request-ID", requestID)

			// Add request ID to context
			ctx := context.WithValue(r.Context(), utils.RequestIDKey, requestID)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
