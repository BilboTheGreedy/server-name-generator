package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bilbothegreedy/server-name-generator/internal/config"
	"github.com/bilbothegreedy/server-name-generator/internal/db"
	"github.com/bilbothegreedy/server-name-generator/internal/handlers"
	"github.com/bilbothegreedy/server-name-generator/internal/middleware"
	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/services"
	"github.com/bilbothegreedy/server-name-generator/internal/session"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
	
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := utils.NewLogger(cfg.LogLevel)
	logger.Info("Starting server name generator service")
	
	// Capture application start time
	startTime := time.Now()

	// Connect to database
	database, err := db.Connect(cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer database.Close()
	
	// Create template cache
	templateCache, err := handlers.CreateTemplateCache()
	if err != nil {
		logger.Fatal("Failed to create template cache", "error", err)
	}
	
	// Initialize session manager
	sessionManager := session.NewSessionManager(database)
	
	// Initialize models
	sequenceModel := models.NewSequenceModel(database)
	reservationModel := models.NewReservationModel(database)
	userModel := models.NewUserModel(database)
	
	// Initialize services
	nameService := services.NewNameGeneratorService(database, sequenceModel, reservationModel, logger)
	
	// Initialize handlers
	app := &handlers.Application{
		Config:         cfg,
		Logger:         logger,
		DB:             database,
		TemplateCache:  templateCache,
		SessionManager: sessionManager,
		NameService:    nameService,
		UserModel:      userModel,
	}
	
	// Create router
	r := chi.NewRouter()
	
	// Middleware
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Timeout(30 * time.Second))
	r.Use(sessionManager.LoadAndSave)
	
	// CSRF protection for web forms
	r.Use(func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(next)
		csrfHandler.SetBaseCookie(http.Cookie{
			HttpOnly: true,
			Path:     "/",
			Secure:   cfg.Environment == "production",
			SameSite: http.SameSiteLaxMode,
		})
		return csrfHandler
	})
	
	// Static files
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))
	
	// API routes (no CSRF)
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.JSONContentType)
		
		// Public API endpoints
		r.Post("/reserve", app.APIReserve)
		r.Post("/commit", app.APICommit)
		
		// Protected API endpoints
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(sessionManager))
			r.Get("/reservations", app.APIGetReservations)
			r.Delete("/reservations/{id}", app.APIDeleteReservation)
			r.Post("/release", app.APIRelease)
			r.Get("/stats", app.APIGetStats)
		})
	})
	
	// Web routes (with CSRF)
	r.Get("/", app.Home)
	r.Get("/login", app.LoginForm)
	r.Post("/login", app.Login)
	r.Get("/logout", app.Logout)
	
	// Admin routes
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.RequireAuth(sessionManager))
		
		r.Get("/", app.AdminDashboard)
		r.Get("/generate", app.AdminGenerate)
		r.Get("/manage", app.AdminManage)
		r.Get("/users", app.AdminUsers)
		r.Post("/users", app.AdminCreateUser)
		r.Get("/users/{id}/edit", app.AdminEditUser)
		r.Post("/users/{id}", app.AdminUpdateUser)
		r.Post("/users/{id}/delete", app.AdminDeleteUser)
	})
	
	// Configure HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in a goroutine
	go func() {
		logger.Info("Server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", "error", err)
		}
	}()
	
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}
	
	logger.Info("Server exited properly")
}