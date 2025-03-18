package session

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
)

// SessionManager is a wrapper around scs.SessionManager
type SessionManager struct {
	*scs.SessionManager
}

// NewSessionManager creates a new session manager
func NewSessionManager(db *sql.DB) *SessionManager {
	// Create a new session manager
	sessionManager := scs.New()
	
	// Configure the session lifetime
	sessionManager.Lifetime = 24 * time.Hour
	
	// Use PostgreSQL as the session store
	sessionManager.Store = postgresstore.New(db)
	
	// Configure cookie security settings
	sessionManager.Cookie.Secure = true    // Set to true in production
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	
	return &SessionManager{
		SessionManager: sessionManager,
	}
}

// GetUserID gets the user ID from the session
func (sm *SessionManager) GetUserID(r *http.Request) string {
	return sm.GetString(r.Context(), "userID")
}

// SetUserID sets the user ID in the session
func (sm *SessionManager) SetUserID(r *http.Request, w http.ResponseWriter, userID string) {
	sm.Put(r.Context(), "userID", userID)
}

// GetUserRole gets the user role from the session
func (sm *SessionManager) GetUserRole(r *http.Request) string {
	return sm.GetString(r.Context(), "userRole")
}

// SetUserRole sets the user role in the session
func (sm *SessionManager) SetUserRole(r *http.Request, w http.ResponseWriter, role string) {
	sm.Put(r.Context(), "userRole", role)
}

// SetFlash sets a flash message
func (sm *SessionManager) SetFlash(r *http.Request, w http.ResponseWriter, message string, messageType string) {
	sm.Put(r.Context(), "flash", message)
	sm.Put(r.Context(), "flashType", messageType)
}

// GetFlash gets and clears the flash message
func (sm *SessionManager) GetFlash(r *http.Request) (string, string) {
	message := sm.PopString(r.Context(), "flash")
	messageType := sm.PopString(r.Context(), "flashType")
	return message, messageType
}

// IsAuthenticated checks if the user is authenticated
func (sm *SessionManager) IsAuthenticated(r *http.Request) bool {
	return sm.GetString(r.Context(), "userID") != ""
}

// IsAdmin checks if the user is an admin
func (sm *SessionManager) IsAdmin(r *http.Request) bool {
	return sm.GetString(r.Context(), "userRole") == "admin"
}

// Logout clears the session
func (sm *SessionManager) Logout(r *http.Request, w http.ResponseWriter) {
	sm.Destroy(r.Context())
}