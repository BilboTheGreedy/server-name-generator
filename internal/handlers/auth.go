package handlers

import (
	"net/http"

	"github.com/bilbothegreedy/server-name-generator/internal/models"
)

// LoginForm handles displaying the login form
func (app *Application) LoginForm(w http.ResponseWriter, r *http.Request) {
	// If already authenticated, redirect to admin dashboard
	if app.SessionManager.IsAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Check if there are users in the system
	users, err := app.UserModel.GetAll(r.Context())
	if err != nil {
		app.Logger.Error("Error checking for users", "error", err)
	}

	// Prepare data
	data := map[string]interface{}{
		"UseDefaultCredentials": len(users) == 0 || hasDefaultAdmin(users),
	}

	// Render login template
	app.RenderTemplate(w, r, "login.tmpl", data)
}

// Login handles the login form submission
func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	// If already authenticated, redirect to admin dashboard
	if app.SessionManager.IsAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Parse form
	err := r.ParseForm()
	if err != nil {
		app.Logger.Error("Error parsing login form", "error", err)
		app.SessionManager.SetFlash(r, w, "There was a problem processing your request", "error")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get form data
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validate form data
	if username == "" || password == "" {
		app.SessionManager.SetFlash(r, w, "Please enter both username and password", "error")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Authenticate user
	user, err := app.UserModel.Authenticate(r.Context(), username, password)
	if err != nil {
		app.Logger.Info("Failed login attempt", "username", username, "error", err)
		app.SessionManager.SetFlash(r, w, "Invalid username or password", "error")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// No user found with those credentials
	if user == nil {
		app.Logger.Info("Failed login attempt - user not found", "username", username)
		app.SessionManager.SetFlash(r, w, "Invalid username or password", "error")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Create a session for the user
	app.SessionManager.SetUserID(r, w, user.ID)
	app.SessionManager.SetUserRole(r, w, user.Role)

	// Log the successful login
	app.Logger.Info("User logged in successfully", "username", user.Username, "id", user.ID)

	// Set success flash message
	app.SessionManager.SetFlash(r, w, "You have been logged in successfully", "success")
	
	// Redirect to the admin dashboard
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Logout handles logging the user out
func (app *Application) Logout(w http.ResponseWriter, r *http.Request) {
	// Destroy the session
	app.SessionManager.Logout(r, w)

	// Set flash message
	app.SessionManager.SetFlash(r, w, "You have been logged out successfully", "success")

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Helper function to check if the default admin user exists
func hasDefaultAdmin(users []*models.User) bool {
	for _, user := range users {
		if user.Username == "admin" {
			return true
		}
	}
	return false
}