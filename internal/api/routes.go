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
	custommw "github.com/bilbothegreedy/server-name-generator/internal/api/middleware"
	"github.com/bilbothegreedy/server-name-generator/internal/auth"
	"github.com/bilbothegreedy/server-name-generator/internal/config"
	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/services"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// SetupRouter configures and returns the API router
func SetupRouter(cfg *config.Config, db *sql.DB, logger *utils.Logger) http.Handler {
	// Initialize models
	sequenceModel := models.NewSequenceModel(db)
	reservationModel := models.NewReservationModel(db)
	userModel := models.NewUserModel(db)

	// Initialize services
	nameService := services.NewNameGeneratorService(db, sequenceModel, reservationModel, logger)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.TokenDuration)

	// Initialize handlers
	reservationHandler := handlers.NewReservationHandler(nameService, logger)
	commitHandler := handlers.NewCommitHandler(nameService, logger)
	releaseHandler := handlers.NewReleaseHandler(nameService, logger)
	authHandler := handlers.NewAuthHandler(userModel, jwtManager, logger)

	// Create router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(custommw.ErrorHandler(logger))
	r.Use(custommw.RequestLogger(logger))
	r.Use(middleware.Timeout(30 * time.Second)) // 30 second timeout

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not readily exceeded by browsers
	}))

	// Public API routes - no authentication required
	r.Route("/api", func(r chi.Router) {
		// Health check endpoint
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{
				"status": "ok",
				"time":   time.Now().UTC().Format(time.RFC3339),
			})
		})

		// Authentication routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/register", authHandler.Register)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(custommw.Authenticate(jwtManager, logger))
				r.Get("/me", authHandler.GetCurrentUser)
			})
		})

		// Protected API routes - require authentication
		r.Group(func(r chi.Router) {
			// Apply authentication middleware
			r.Use(custommw.Authenticate(jwtManager, logger))

			// Regular user endpoints
			r.With(custommw.ValidateReservationRequest(logger)).
				Post("/reserve", reservationHandler.Reserve)

			r.With(custommw.ValidateCommitRequest(logger)).
				Post("/commit", commitHandler.Commit)

			// Admin-only endpoints
			r.Group(func(r chi.Router) {
				r.Use(custommw.RequireRole(models.RoleAdmin))

				// Get all reservations
				r.Get("/reservations", func(w http.ResponseWriter, r *http.Request) {
					reservations, err := nameService.GetAllReservations(r.Context())
					if err != nil {
						logger.Error("Failed to get reservations", "error", err)
						utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get reservations")
						return
					}
					utils.RespondWithJSON(w, http.StatusOK, reservations)
				})

				// Release endpoint - requires admin role
				r.With(custommw.ValidateReleaseRequest(logger)).
					Post("/release", releaseHandler.Release)

				// Delete a reservation (admin use only)
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

						// Check for not found errors
						if strings.Contains(err.Error(), "not found") {
							utils.RespondWithError(w, http.StatusNotFound, "Reservation not found")
							return
						}

						utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete reservation: "+err.Error())
						return
					}

					logger.Info("Reservation deleted successfully", "id", id)

					// Return success response
					utils.RespondWithJSON(w, http.StatusOK, map[string]string{
						"message": "Reservation deleted successfully",
					})
				})

				// Get stats for dashboard
				r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
					stats, err := nameService.GetStats(r.Context())
					if err != nil {
						logger.Error("Failed to get stats", "error", err)
						utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get stats")
						return
					}
					utils.RespondWithJSON(w, http.StatusOK, stats)
				})
			})
		})
	})

	// Redirects from root and /admin/ to /admin
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	})

	r.Get("/admin/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	})

	// Serve admin dashboard with authentication
	r.Group(func(r chi.Router) {
		// Use optional authentication to allow access to login page
		r.Use(custommw.OptionalAuth(jwtManager, logger))

		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./static/admin.html")
		})

		// Serve static files for the admin dashboard
		fileServer := http.FileServer(http.Dir("./static"))
		r.Handle("/static/*", http.StripPrefix("/static", fileServer))
	})

	return r
}
