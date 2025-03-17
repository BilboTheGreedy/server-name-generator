package utils

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

// Logger is a wrapper around slog.Logger with predefined methods
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new logger with the specified level
func NewLogger(level string) *Logger {
	// Set log level based on configuration
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create logger options
	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format timestamps as RFC3339
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.Attr{
						Key:   slog.TimeKey,
						Value: slog.StringValue(t.Format(time.RFC3339)),
					}
				}
			}
			return a
		},
	}

	// Create JSON handler
	handler := slog.NewJSONHandler(os.Stdout, opts)

	// Create the logger
	logger := slog.New(handler)

	return &Logger{logger}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...any) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
}

// WithContext creates a new logger with additional context
func (l *Logger) WithContext(ctx ...any) *Logger {
	return &Logger{l.Logger.With(ctx...)}
}

// LogRequest logs information about an HTTP request
func (l *Logger) LogRequest(method, path, remoteAddr, userAgent string) {
	l.Info("Request received",
		"method", method,
		"path", path,
		"remote_addr", remoteAddr,
		"user_agent", userAgent,
	)
}

// LogResponse logs information about an HTTP response
func (l *Logger) LogResponse(method, path string, statusCode int, duration time.Duration) {
	l.Info("Response sent",
		"method", method,
		"path", path,
		"status", statusCode,
		"duration_ms", duration.Milliseconds(),
	)
}

// LogError logs an error with its stack trace
func (l *Logger) LogError(err error, msg string, args ...any) {
	if err == nil {
		return
	}

	// Add the error to the args list
	newArgs := append([]any{"error", err.Error()}, args...)
	l.Error(msg, newArgs...)
}
