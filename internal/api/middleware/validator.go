package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// ValidateReservationRequest validates the incoming reservation request
func ValidateReservationRequest(logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate POST requests
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Check content type
			if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
				logger.Warn("Invalid Content-Type", "contentType", contentType)
				utils.RespondWithError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
				return
			}

			// Parse the request body
			var payload models.ReservationPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				logger.Error("Failed to decode request body", "error", err)
				utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}

			// Important: Reset the request body for the next handler
			r.Body.Close()
			r.Body = utils.CreateReadCloser(payload)

			// Proceed to the next handler - we're no longer validating required fields
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateCommitRequest validates the incoming commit request
func ValidateCommitRequest(logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate POST requests
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Check content type
			if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
				logger.Warn("Invalid Content-Type", "contentType", contentType)
				utils.RespondWithError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
				return
			}

			// Parse and validate the request body
			var payload models.CommitPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				logger.Error("Failed to decode commit request body", "error", err)
				utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}

			// Important: Reset the request body for the next handler
			r.Body.Close()
			r.Body = utils.CreateReadCloser(payload)

			// Validate the payload (still required for commit)
			if err := utils.Validate(payload); err != nil {
				logger.Error("Invalid commit payload", "error", err)
				utils.RespondWithError(w, http.StatusBadRequest, err.Error())
				return
			}

			// Proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateReleaseRequest validates the incoming release request
func ValidateReleaseRequest(logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate POST requests
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Check content type
			if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
				logger.Warn("Invalid Content-Type", "contentType", contentType)
				utils.RespondWithError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
				return
			}

			// Parse and validate the request body
			var payload models.ReleasePayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				logger.Error("Failed to decode release request body", "error", err)
				utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}

			// Important: Reset the request body for the next handler
			r.Body.Close()
			r.Body = utils.CreateReadCloser(payload)

			// Validate the payload
			if err := utils.Validate(payload); err != nil {
				logger.Error("Invalid release payload", "error", err)
				utils.RespondWithError(w, http.StatusBadRequest, err.Error())
				return
			}

			// Proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
