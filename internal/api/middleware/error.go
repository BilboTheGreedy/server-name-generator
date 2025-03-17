package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// ErrorHandler middleware to catch panics and handle errors gracefully
func ErrorHandler(logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the stack trace
					logger.Error(
						"Panic recovered",
						"error", err,
						"stacktrace", string(debug.Stack()),
						"path", r.URL.Path,
						"method", r.Method,
					)

					// Respond with error
					var errorMsg string
					switch e := err.(type) {
					case string:
						errorMsg = e
					case error:
						errorMsg = e.Error()
					default:
						errorMsg = "Unknown server error"
					}

					utils.RespondWithError(w, http.StatusInternalServerError, errorMsg)
				}
			}()

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs incoming requests
func RequestLogger(logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Request received",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Create a custom response writer to capture status code
			rw := utils.NewResponseWriter(w)

			// Call the next handler
			next.ServeHTTP(rw, r)

			// Log the response status
			logger.Info("Response sent",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.StatusCode,
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
