package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/bilbothegreedy/server-name-generator/internal/config"
	"github.com/bilbothegreedy/server-name-generator/internal/models"
	"github.com/bilbothegreedy/server-name-generator/internal/services"
	"github.com/bilbothegreedy/server-name-generator/internal/session"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
	"github.com/justinas/nosurf"
)

// Application holds the application-wide dependencies
type Application struct {
	Config         *config.Config
	Logger         *utils.Logger
	DB             *sql.DB
	TemplateCache  map[string]*template.Template
	SessionManager *session.SessionManager
	NameService    *services.NameGeneratorService
	UserModel      *models.UserModel
}

// TemplateData holds data passed to templates
type TemplateData struct {
	CSRFToken       string
	Flash           string
	FlashType       string
	IsAuthenticated bool
	IsAdmin         bool
	User            *models.User
	Data            map[string]interface{}
	CurrentYear     int
}

// CreateTemplateCache creates a cache of templates
func CreateTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Find all page templates
	pages, err := filepath.Glob("./internal/templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Find all admin templates
	adminPages, err := filepath.Glob("./internal/templates/admin/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Find all auth templates
	authPages, err := filepath.Glob("./internal/templates/auth/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Combine all templates
	pages = append(pages, adminPages...)
	pages = append(pages, authPages...)

	// Create template cache for each page
	for _, page := range pages {
		name := filepath.Base(page)

		// Parse the page template
		ts, err := template.New(name).Funcs(template.FuncMap{
			"formatDate": formatDate,
			"inc":        increment,
			"safe":       safe,
			"mod":        mod,
			"eq":         equals,
		}).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Add base layout
		ts, err = ts.ParseGlob("./internal/templates/layouts/*.tmpl")
		if err != nil {
			return nil, err
		}

		// Add partials
		ts, err = ts.ParseGlob("./internal/templates/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

// AddDefaultData adds default data to the template data
func (app *Application) AddDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	if td == nil {
		td = &TemplateData{}
	}
	
	// Add CSRF token
	td.CSRFToken = nosurf.Token(r)
	
	// Add flash message
	flash, flashType := app.SessionManager.GetFlash(r)
	td.Flash = flash
	td.FlashType = flashType
	
	// Add authentication status
	td.IsAuthenticated = app.SessionManager.IsAuthenticated(r)
	
	// Add admin status
	td.IsAdmin = app.SessionManager.IsAdmin(r)
	
	// Add current year for copyright
	td.CurrentYear = time.Now().Year()
	
	// If authenticated, get user
	if td.IsAuthenticated {
		userID := app.SessionManager.GetUserID(r)
		user, err := app.UserModel.GetByID(r.Context(), userID)
		if err == nil && user != nil {
			td.User = user
		}
	}
	
	// Initialize Data map if not set
	if td.Data == nil {
		td.Data = make(map[string]interface{})
	}
	
	return td
}

// RenderTemplate renders a template with the given data
func (app *Application) RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, data map[string]interface{}) {
	var ts *template.Template
	var err error
	
	// Check if template is in cache
	if ts, ok := app.TemplateCache[tmpl]; ok {
		ts = ts
	} else {
		app.Logger.Error("Template not found in cache", "template", tmpl)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	// Create a buffer to store the rendered template
	buf := new(bytes.Buffer)
	
	// Add default data
	td := app.AddDefaultData(&TemplateData{Data: data}, r)
	
	// Execute the template
	err = ts.Execute(buf, td)
	if err != nil {
		app.Logger.Error("Error executing template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	// Write the rendered template to the response
	_, err = buf.WriteTo(w)
	if err != nil {
		app.Logger.Error("Error writing template to response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// ErrorResponse sends a JSON error response
func (app *Application) ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":"%s"}`, message)
}

// JSONResponse sends a JSON response
func (app *Application) JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	utils.WriteJSON(w, data)
}

// Template helper functions
func formatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func increment(n int) int {
	return n + 1
}

func safe(s string) template.HTML {
	return template.HTML(s)
}

func mod(a, b int) int {
	return a % b
}

func equals(a, b interface{}) bool {
	return a == b
}