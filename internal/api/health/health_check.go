package health

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/bilbothegreedy/server-name-generator/internal/config"
	"github.com/bilbothegreedy/server-name-generator/internal/utils"
)

// HealthCheckResponse represents the structure of the health check response.
type HealthCheckResponse struct {
	Status     string            `json:"status"`
	Timestamp  time.Time         `json:"timestamp"`
	Version    string            `json:"version"`
	Uptime     string            `json:"uptime"`
	GoVersion  string            `json:"goVersion"`
	Database   DatabaseStatus    `json:"database"`
	System     SystemStatus      `json:"system"`
	Deployment DeploymentDetails `json:"deployment"`
	Components map[string]string `json:"components"`
}

// DatabaseStatus represents the status of the database connection.
type DatabaseStatus struct {
	Connected  bool   `json:"connected"`
	Error      string `json:"error,omitempty"`
	Migrations string `json:"migrations,omitempty"`
}

// SystemStatus represents system-related health information.
type SystemStatus struct {
	OSType        string `json:"osType"`
	Arch          string `json:"arch"`
	CPUs          int    `json:"cpus"`
	GoroutinesNum int    `json:"goroutines"`
	TotalAllocMB  int    `json:"totalAllocMB"`
	NumCPU        int    `json:"numCPU"`
	NumGoRoutine  int    `json:"numGoRoutine"`
}

// DeploymentDetails provides information about the application deployment.
type DeploymentDetails struct {
	ExecutablePath string `json:"executablePath"`
	WorkingDir     string `json:"workingDir"`
	Hostname       string `json:"hostname"`
}

// GetHealthCheck returns an HTTP handler for the health check endpoint.
func GetHealthCheck(cfg *config.Config, db *sql.DB, logger *utils.Logger, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity.
		dbStatus := checkDatabaseStatus(db)

		// Get runtime metrics.
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Get deployment details.
		execPath, _ := os.Executable()
		workDir, _ := os.Getwd()
		hostname, _ := os.Hostname()

		response := HealthCheckResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC(),
			Version:   getAppVersion(),
			Uptime:    time.Since(startTime).String(),
			GoVersion: runtime.Version(),
			Database:  dbStatus,
			System: SystemStatus{
				OSType:        runtime.GOOS,
				Arch:          runtime.GOARCH,
				CPUs:          runtime.NumCPU(),
				GoroutinesNum: runtime.NumGoroutine(),
				TotalAllocMB:  int(m.TotalAlloc / 1024 / 1024),
				NumCPU:        runtime.NumCPU(),
				NumGoRoutine:  runtime.NumGoroutine(),
			},
			Deployment: DeploymentDetails{
				ExecutablePath: execPath,
				WorkingDir:     workDir,
				Hostname:       hostname,
			},
			Components: map[string]string{
				"database":   dbStatus.Status(),
				"env":        cfg.LogLevel,
				"migrations": dbStatus.Migrations,
			},
		}

		// If database is not connected, mark overall status as degraded.
		if !dbStatus.Connected {
			response.Status = "degraded"
		}

		// Send response.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func checkDatabaseStatus(db *sql.DB) DatabaseStatus {
	status := DatabaseStatus{Connected: false}
	// Ping database.
	if err := db.Ping(); err != nil {
		status.Error = err.Error()
		return status
	}

	status.Connected = true

	// Check migration status (optional).
	migrationPath := findMigrationPath()
	if migrationPath != "" {
		status.Migrations = fmt.Sprintf("Migrations found at: %s", migrationPath)
	} else {
		status.Migrations = "No migrations directory found"
	}

	return status
}

func findMigrationPath() string {
	possiblePaths := []string{
		// Try executable directory.
		func() string {
			execPath, err := os.Executable()
			if err == nil {
				return filepath.Join(filepath.Dir(execPath), "migrations")
			}
			return ""
		}(),
		// Try current working directory.
		func() string {
			wd, err := os.Getwd()
			if err == nil {
				return filepath.Join(wd, "migrations")
			}
			return ""
		}(),
		// Try relative to executable.
		func() string {
			execPath, err := os.Executable()
			if err == nil {
				return filepath.Join(filepath.Dir(execPath), "..", "migrations")
			}
			return ""
		}(),
		"./migrations",
		"../migrations",
		"../../migrations",
	}

	for _, path := range possiblePaths {
		if path == "" {
			continue
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	return ""
}

// Status provides a human-readable status for the database.
func (d DatabaseStatus) Status() string {
	if d.Connected {
		return "healthy"
	}
	return "unavailable"
}

func getAppVersion() string {
	// In a real application, this could be set at build time.
	return "1.0.0"
}
