package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Home handles the home page
func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	// Redirect to admin dashboard if authenticated
	if app.SessionManager.IsAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	
	// Otherwise redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// AdminDashboard handles the admin dashboard page
func (app *Application) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	// Get stats for dashboard
	stats, err := app.NameService.GetStats(r.Context())
	if err != nil {
		app.Logger.Error("Error getting stats", "error", err)
		app.SessionManager.SetFlash(r, w, "Error loading dashboard data", "error")
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"CurrentPage":        "dashboard",
		"Stats":              stats,
		"RecentReservations": stats.RecentReservations,
	}

	// Render template
	app.RenderTemplate(w, r, "dashboard.tmpl", data)
}

// AdminGenerate handles the name generation page
func (app *Application) AdminGenerate(w http.ResponseWriter, r *http.Request) {
	// Prepare template data
	data := map[string]interface{}{
		"CurrentPage": "generate",
	}

	// Render template
	app.RenderTemplate(w, r, "generate.tmpl", data)
}

// AdminManage handles the reservation management page
func (app *Application) AdminManage(w http.ResponseWriter, r *http.Request) {
	// Get all reservations
	reservations, err := app.NameService.GetAllReservations(r.Context())
	if err != nil {
		app.Logger.Error("Error getting reservations", "error", err)
		app.SessionManager.SetFlash(r, w, "Error loading reservations", "error")
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"CurrentPage":   "manage",
		"Reservations":  reservations,
	}

	// Render template
	app.RenderTemplate(w, r, "manage.tmpl", data)
}

// AdminReservationsPartial handles the AJAX request for reservations table body
func (app *Application) AdminReservationsPartial(w http.ResponseWriter, r *http.Request) {
	// Get all reservations
	reservations, err := app.NameService.GetAllReservations(r.Context())
	if err != nil {
		app.Logger.Error("Error getting reservations", "error", err)
		http.Error(w, "Failed to load reservations", http.StatusInternalServerError)
		return
	}

	// Prepare data
	data := map[string]interface{}{
		"Reservations": reservations,
	}

	// Render just the reservations partial
	app.RenderTemplate(w, r, "reservations_partial.tmpl", data)
}

// AdminUsers handles the user management page
func (app *Application) AdminUsers(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !app.SessionManager.IsAdmin(r) {
		app.SessionManager.SetFlash(r, w, "You don't have permission to access this page", "error")
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Get all users
	users, err := app.UserModel.GetAll(r.Context())
	if err != nil {
		app.Logger.Error("Error getting users", "error", err)
		app.SessionManager.SetFlash(r, w, "Error loading users", "error")
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"CurrentPage": "users",
		"Users":       users,
	}

	// Render template
	app.RenderTemplate(w, r, "users.tmpl", data)
}

// AdminCreateUser handles creating a new user
func (app *Application) AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !app.SessionManager.IsAdmin(r) {
		app.SessionManager.SetFlash(r, w, "You don't have permission to perform this action", "error")
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Parse form
	err := r.ParseForm()
	if err != nil {
		app.Logger.Error("Error parsing form", "error", err)
		app.SessionManager.SetFlash(r, w, "Error processing form", "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// Get form data
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	role := r.FormValue("role")

	// Validate form data
	if username == "" || email == "" || password == "" || role == "" {
		app.SessionManager.SetFlash(r, w, "All fields are required", "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// Create user
	err = app.UserModel.CreateUser(r.Context(), username, email, password, role)
	if err != nil {
		app.Logger.Error("Error creating user", "error", err)
		app.SessionManager.SetFlash(r, w, "Error creating user: "+err.Error(), "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// Set success flash message
	app.SessionManager.SetFlash(r, w, "User created successfully", "success")
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// AdminEditUser handles the edit user page
func (app *Application) AdminEditUser(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !app.SessionManager.IsAdmin(r) {
		app.SessionManager.SetFlash(r, w, "You don't have permission to access this page", "error")
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Get user ID from URL
	userID := chi.URLParam(r, "id")
	if userID == "" {
		app.SessionManager.SetFlash(r, w, "Invalid user ID", "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// Get user from database
	user, err := app.UserModel.GetByID(r.Context(), userID)
	if err != nil {
		app.Logger.Error("Error getting user", "error", err)
		app.SessionManager.SetFlash(r, w, "Error loading user", "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	if user == nil {
		app.SessionManager.SetFlash(r, w, "User not found", "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"CurrentPage": "users",
		"User":        user,
	}

	// Render template
	app.RenderTemplate(w, r, "edit_user.tmpl", data)
}

// AdminUpdateUser handles updating a user
func (app *Application) AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !app.SessionManager.IsAdmin(r) {
		app.SessionManager.SetFlash(r, w, "You don't have permission to perform this action", "error")
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Get user ID from URL
	userID := chi.URLParam(r, "id")
	if userID == "" {
		app.SessionManager.SetFlash(r, w, "Invalid user ID", "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// Parse form
	err := r.ParseForm()
	if err != nil {
		app.Logger.Error("Error parsing form", "error", err)
		app.SessionManager.SetFlash(r, w, "Error processing form", "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// Get form data
	username := r.FormValue("username")
	email := r.FormValue("email")
	role := r.FormValue("role")
	password := r.FormValue("password")

	// Validate form data
	if username == "" || email == "" || role == "" {
		app.SessionManager.SetFlash(r, w, "Username, email, and role are required", "error")
		http.Redirect(w, r, "/admin/users/"+userID+"/edit", http.StatusSeeOther)
		return
	}

	// Update user
	err = app.UserModel.UpdateUser(r.Context(), userID, username, email, role)
	if err != nil {
		app.Logger.Error("Error updating user", "error", err)
		app.SessionManager.SetFlash(r, w, "Error updating user: "+err.Error(), "error")
		http.Redirect(w, r, "/admin/users/"+userID+"/edit", http.StatusSeeOther)
		return
	}

	// Update password if provided
	if password != "" {
		err = app.UserModel.UpdatePassword(r.Context(), userID, password)
		if err != nil {
			app.Logger.Error("Error updating password", "error", err)
			app.SessionManager.SetFlash(r, w, "Error updating password: "+err.Error(), "error")
			http.Redirect(w, r, "/admin/users/"+userID+"/edit", http.StatusSeeOther)
			return
		}
	}

	// Set success flash message
	app.SessionManager.SetFlash(r, w, "User updated successfully", "success")
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// AdminDeleteUser handles deleting a user
func (app *Application) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !app.SessionManager.IsAdmin(r) {
		app.SessionManager.SetFlash(r, w, "You don't have permission to perform this action", "error")
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Get user ID from URL
	userID := chi.URLParam(r, "id")
	if userID == "" {
		app.SessionManager.SetFlash(r, w, "Invalid user ID", "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// Don't allow deleting the current user
	currentUserID := app.SessionManager.GetUserID(r)
	if userID == currentUserID {
		app.SessionManager.SetFlash(r, w, "You cannot delete your own account", "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// Delete user
	err := app.UserModel.Delete(r.Context(), userID)
	if err != nil {
		app.Logger.Error("Error deleting user", "error", err)
		app.SessionManager.SetFlash(r, w, "Error deleting user: "+err.Error(), "error")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	// Set success flash message
	app.SessionManager.SetFlash(r, w, "User deleted successfully", "success")
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}