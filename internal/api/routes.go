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
	r.Use(custommw.TimeoutMiddleware(logger, 30)) // 30 second timeout

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not readily exceeded by browsers
	}))

	// Set a reasonable request body limit
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	// Routes
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
	})

	return r
}
