// internal/errors/errors.go
package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Error types
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeDatabase     ErrorType = "database"
	ErrorTypeNotFound     ErrorType = "not_found"
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	ErrorTypeForbidden    ErrorType = "forbidden"
	ErrorTypeInternal     ErrorType = "internal"
	ErrorTypeConflict     ErrorType = "conflict"
	ErrorTypeBadRequest   ErrorType = "bad_request"
)

// AppError represents an application error with additional context
type AppError struct {
	Type       ErrorType `json:"-"`                   // Error type (not exposed in JSON)
	Message    string    `json:"message"`             // User-friendly error message
	Detail     string    `json:"detail,omitempty"`    // Optional detailed error message
	StatusCode int       `json:"-"`                   // HTTP status code (not exposed in JSON)
	Err        error     `json:"-"`                   // Original error (not exposed in JSON)
	RequestID  string    `json:"requestId,omitempty"` // For tracking in logs
	Code       string    `json:"code,omitempty"`      // Optional error code
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	}
	return e.Message
}

// Unwrap returns the original error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Error constructors for common error types
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

func NewDatabaseError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeDatabase,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

func NewBadRequestError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

// WithRequestID adds request ID to the error
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithDetail adds additional detail to the error
func (e *AppError) WithDetail(detail string) *AppError {
	e.Detail = detail
	return e
}

// WithCode adds an error code to the error
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}
