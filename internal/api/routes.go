package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/bilbothegreedy/server-name-generator/internal/api/handlers"
	custommw "github.com/bilbothegreedy/server-name-generator/internal/api/middleware"
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

	// Initialize services
	nameService := services.NewNameGeneratorService(db, sequenceModel, reservationModel, logger)

	// Initialize handlers
	reservationHandler := handlers.NewReservationHandler(nameService, logger)
	commitHandler := handlers.NewCommitHandler(nameService, logger)

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
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not readily exceeded by browsers
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Health check endpoint
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{
				"status": "ok",
				"time":   time.Now().UTC().Format(time.RFC3339),
			})
		})

		// Reservation endpoint
		r.With(custommw.ValidateReservationRequest(logger)).
			Post("/reserve", reservationHandler.Reserve)

		// Commit endpoint
		r.With(custommw.ValidateCommitRequest(logger)).
			Post("/commit", commitHandler.Commit)

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

		// Delete a reservation (for admin use)
		r.Delete("/reservations/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			if id == "" {
				utils.RespondWithError(w, http.StatusBadRequest, "Missing reservation ID")
				return
			}

			err := nameService.DeleteReservation(r.Context(), id)
			if err != nil {
				logger.Error("Failed to delete reservation", "error", err)
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete reservation: "+err.Error())
				return
			}

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

	// Redirects from root and /admin/ to /admin
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	})

	r.Get("/admin/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	})

	// Serve admin dashboard
	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/admin.html")
	})

	// Serve static files for the admin dashboard
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return r
}
