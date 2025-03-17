// internal/utils/logger.go (update)
package utils

import (
	"context"
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

// WithRequestID adds request ID to logger
func (l *Logger) WithRequestID(ctx context.Context) *Logger {
	logger := l.Logger

	// Add request ID from context if available
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.With("request_id", requestID)
	}

	return &Logger{logger}
}

// WithContext creates a new logger with additional context
func (l *Logger) WithContext(ctx ...any) *Logger {
	return &Logger{l.Logger.With(ctx...)}
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

// LogRequest logs information about an HTTP request
func (l *Logger) LogRequest(ctx context.Context, method, path, remoteAddr, userAgent string) {
	logger := l.WithRequestID(ctx)
	logger.Info("Request received",
		"method", method,
		"path", path,
		"remote_addr", remoteAddr,
		"user_agent", userAgent,
	)
}

// LogResponse logs information about an HTTP response
func (l *Logger) LogResponse(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	logger := l.WithRequestID(ctx)
	logger.Info("Response sent",
		"method", method,
		"path", path,
		"status", statusCode,
		"duration_ms", duration.Milliseconds(),
	)
}

// LogError logs an error with its request ID context
func (l *Logger) LogError(ctx context.Context, err error, msg string, args ...any) {
	if err == nil {
		return
	}

	logger := l.WithRequestID(ctx)

	// Add the error to the args list
	newArgs := append([]any{"error", err.Error()}, args...)
	logger.Error(msg, newArgs...)
}
