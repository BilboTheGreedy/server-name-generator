package api

import (
	"github.com/bilbothegreedy/server-name-generator/internal/api/handlers"
	custommw "github.com/bilbothegreedy/server-name-generator/internal/api/middleware"
	"github.com/bilbothegreedy/server-name-generator/internal/auth"
	"github.com/bilbothegreedy/server-name-generator/internal/config"
	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
	"github.com/go-chi/chi/v5"
)

// SetupAuthRoutes configures and returns authentication routes
func SetupAuthRoutes(r chi.Router, cfg *config.Config, userModel *models.UserModel, logger *utils.Logger) *auth.JWTManager {
	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.TokenDuration)

	// Initialize auth handler
	authHandler := handlers.NewAuthHandler(userModel, jwtManager, logger)

	// Public auth routes (no authentication required)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(custommw.Authenticate(jwtManager, logger))
			r.Get("/me", authHandler.GetCurrentUser)
		})
	})

	// Return the JWT manager and authentication middleware for use in other routes
	return jwtManager
}

// ApplyAuthMiddleware applies authentication middleware to protected routes
func ApplyAuthMiddleware(r chi.Router, jwtManager *auth.JWTManager, logger *utils.Logger, roles ...string) chi.Router {
	// Create a new router with authentication middleware
	authRouter := chi.NewRouter()

	// Apply authentication middleware
	authRouter.Use(custommw.Authenticate(jwtManager, logger))

	// Apply role-based authorization if roles are specified
	if len(roles) > 0 {
		authRouter.Use(custommw.RequireRole(roles...))
	}

	// Mount the authenticated router
	r.Mount("/", authRouter)

	return authRouter
}
