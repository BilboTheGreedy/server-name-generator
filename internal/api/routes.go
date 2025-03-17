package api

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/bilbothegreedy/server-name-generator/internal/api/handlers"
	"github.com/bilbothegreedy/server-name-generator/internal/api/health"
	custommw "github.com/bilbothegreedy/server-name-generator/internal/api/middleware"
	"github.com/bilbothegreedy/server-name-generator/internal/auth"
	"github.com/bilbothegreedy/server-name-generator/internal/config"
	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/services"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// SetupRouter configures and returns the API router.
// The startTime parameter should be the application start time.
func SetupRouter(cfg *config.Config, db *sql.DB, logger *utils.Logger, startTime time.Time) http.Handler {
	// Initialize models.
	sequenceModel := models.NewSequenceModel(db)
	reservationModel := models.NewReservationModel(db)
	userModel := models.NewUserModel(db)
	apiKeyModel := models.NewAPIKeyModel(db)

	// Initialize services.
	nameService := services.NewNameGeneratorService(db, sequenceModel, reservationModel, logger)

	// Initialize JWT manager.
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.TokenDuration)

	// Initialize handlers.
	reservationHandler := handlers.NewReservationHandler(nameService, logger)
	commitHandler := handlers.NewCommitHandler(nameService, logger)
	releaseHandler := handlers.NewReleaseHandler(nameService, logger)
	authHandler := handlers.NewAuthHandler(userModel, jwtManager, logger)
	userManagementHandler := handlers.NewUserManagementHandler(userModel, logger)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyModel, userModel, logger)

	// Create router.
	r := chi.NewRouter()

	// Global middleware.
	r.Use(custommw.RequestIDMiddleware()) // Custom request ID middleware.
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID) // Fallback request ID.
	r.Use(middleware.Recoverer) // Fallback recovery.
	r.Use(custommw.ErrorHandler(logger))
	r.Use(custommw.RequestLogger(logger))
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS configuration.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public API routes – no authentication required.
	r.Route("/api", func(r chi.Router) {
		// Authentication routes.
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/register", authHandler.Register)

			// Protected auth routes.
			r.Group(func(r chi.Router) {
				r.Use(custommw.CombinedAuth(jwtManager, apiKeyModel, logger))
				r.Get("/me", authHandler.GetCurrentUser)
			})
		})

		// Protected API routes – require authentication.
		r.Group(func(r chi.Router) {
			// Combined authentication middleware (JWT or API Key).
			r.Use(custommw.CombinedAuth(jwtManager, apiKeyModel, logger))

			// Regular user endpoints.
			r.With(custommw.ValidateReservationRequest(logger)).
				Post("/reserve", reservationHandler.Reserve)

			r.With(custommw.ValidateCommitRequest(logger)).
				Post("/commit", commitHandler.Commit)

			// Admin-only endpoints.
			r.Group(func(r chi.Router) {
				r.Use(custommw.RequireRole(models.RoleAdmin))

				// Get all reservations.
				r.Get("/reservations", func(w http.ResponseWriter, r *http.Request) {
					reservations, err := nameService.GetAllReservations(r.Context())
					if err != nil {
						logger.Error("Failed to get reservations", "error", err)
						utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get reservations")
						return
					}
					utils.RespondWithJSON(w, http.StatusOK, reservations)
				})

				// Release endpoint – requires admin role.
				r.With(custommw.ValidateReleaseRequest(logger)).
					Post("/release", releaseHandler.Release)

				// Delete a reservation (admin use only).
				r.Delete("/reservations/{id}", func(w http.ResponseWriter, r *http.Request) {
					id := chi.URLParam(r, "id")
					if id == "" {
						logger.Error("Missing reservation ID in delete request", "path", r.URL.Path)
						utils.RespondWithError(w, http.StatusBadRequest, "Missing reservation ID")
						return
					}

					logger.Info("Attempting to delete reservation", "id", id)
					err := nameService.DeleteReservation(r.Context(), id)
					if err != nil {
						logger.Error("Failed to delete reservation", "error", err, "id", id)
						if strings.Contains(err.Error(), "not found") {
							utils.RespondWithError(w, http.StatusNotFound, "Reservation not found")
							return
						}
						utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete reservation: "+err.Error())
						return
					}

					logger.Info("Reservation deleted successfully", "id", id)
					utils.RespondWithJSON(w, http.StatusOK, map[string]string{
						"message": "Reservation deleted successfully",
					})
				})

				// Get stats for dashboard.
				r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
					stats, err := nameService.GetStats(r.Context())
					if err != nil {
						logger.Error("Failed to get stats", "error", err)
						utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get stats")
						return
					}
					utils.RespondWithJSON(w, http.StatusOK, stats)
				})

				// User management endpoints (admin only).
				r.Route("/users", func(r chi.Router) {
					r.Get("/", userManagementHandler.GetAllUsers)
					r.Post("/", userManagementHandler.CreateUser)
					r.Get("/{id}", userManagementHandler.GetUser)
					r.Put("/{id}", userManagementHandler.UpdateUser)
					r.Post("/{id}/password", userManagementHandler.ChangeUserPassword)
					r.Delete("/{id}", userManagementHandler.DeleteUser)
				})

				// Admin API key management.
				r.Get("/api-keys", apiKeyHandler.GetAll)
			})

			// User API key management (for current user).
			r.Route("/api-keys", func(r chi.Router) {
				r.Get("/", apiKeyHandler.GetAllForUser)
				r.Post("/", apiKeyHandler.Create)
				r.Delete("/{id}", apiKeyHandler.Revoke)
			})
		})
	})

	// Redirects from root and /admin/ to /admin.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	})
	r.Get("/admin/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	})

	// Serve admin dashboard with authentication.
	r.Group(func(r chi.Router) {
		r.Use(custommw.OptionalAuth(jwtManager, logger))
		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./static/admin.html")
		})
		fileServer := http.FileServer(http.Dir("./static"))
		r.Handle("/static/*", http.StripPrefix("/static", fileServer))
	})

	// Register the comprehensive health check endpoint.
	r.Get("/api/health", health.GetHealthCheck(cfg, db, logger, startTime))
	return r
}
