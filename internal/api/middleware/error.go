// internal/api/middleware/error.go (update)
package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	apperrors "github.com/bilbothegreedy/server-name-generator/internal/errors"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// ErrorHandler middleware to catch panics and handle errors gracefully
func ErrorHandler(logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					// Get stack trace
					stack := debug.Stack()

					// Log the panic with context
					ctx := r.Context()
					logger.WithContext(ctx).Error(
						"Panic recovered",
						"error", rec,
						"stacktrace", string(stack),
						"path", r.URL.Path,
						"method", r.Method,
					)

					// Convert to error message
					var errorMsg string
					switch e := rec.(type) {
					case string:
						errorMsg = e
					case error:
						errorMsg = e.Error()
					default:
						errorMsg = "Unknown server error"
					}

					// Create application error
					appErr := apperrors.NewInternalError("Server error", nil).
						WithDetail(errorMsg)

					// Add request ID if present
					if requestID, ok := ctx.Value(utils.RequestIDKey).(string); ok {
						appErr.WithRequestID(requestID)
					}

					// Respond with error
					utils.RespondWithAppError(w, ctx, appErr)
				}
			}()

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs each request and its response
func RequestLogger(logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := r.Context()

			// Log request
			logger.LogRequest(
				ctx,
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				r.UserAgent(),
			)

			// Create a custom response writer to capture status code
			rw := utils.NewResponseWriter(w)

			// Call the next handler
			next.ServeHTTP(rw, r)

			// Log response with duration
			duration := time.Since(start)
			logger.LogResponse(
				ctx,
				r.Method,
				r.URL.Path,
				rw.StatusCode,
				duration,
			)
		})
	}
}

// TimeoutMiddleware adds a timeout to the request context
func TimeoutMiddleware(logger *utils.Logger, timeout int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new context with timeout
			ctx, cancel := utils.ContextWithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal when the handler is done
			done := make(chan struct{})

			// Use the new context in the request
			r = r.WithContext(ctx)

			go func() {
				next.ServeHTTP(w, r)
				done <- struct{}{}
			}()

			select {
			case <-done:
				// Handler finished normally
				return
			case <-ctx.Done():
				// Context timed out
				if errors.Is(ctx.Err(), utils.ErrRequestTimeout) {
					logger.Warn("Request timed out",
						"method", r.Method,
						"path", r.URL.Path,
						"timeout", timeout,
					)
					utils.RespondWithError(w, http.StatusGatewayTimeout, "Request timed out")
				}
				return
			}
		})
	}
}
