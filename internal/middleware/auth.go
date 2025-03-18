package middleware

import (
	"net/http"

	"github.com/bilbothegreedy/server-name-generator/internal/session"
)

// RequireAuth middleware ensures a user is authenticated
func RequireAuth(sessionManager *session.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is authenticated
			if !sessionManager.IsAuthenticated(r) {
				// Set flash message
				sessionManager.SetFlash(r, w, "You must be logged in to access this page", "error")
				
				// Redirect to login page
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			
			// User is authenticated, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin middleware ensures a user is an admin
func RequireAdmin(sessionManager *session.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is authenticated and an admin
			if !sessionManager.IsAuthenticated(r) {
				// Set flash message
				sessionManager.SetFlash(r, w, "You must be logged in to access this page", "error")
				
				// Redirect to login page
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			
			// Check if user is an admin
			if !sessionManager.IsAdmin(r) {
				// Set flash message
				sessionManager.SetFlash(r, w, "You don't have permission to access this page", "error")
				
				// Redirect to dashboard
				http.Redirect(w, r, "/admin", http.StatusSeeOther)
				return
			}
			
			// User is an admin, continue
			next.ServeHTTP(w, r)
		})
	}
}

// JSONContentType middleware sets the content type to application/json
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}