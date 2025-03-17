// internal/utils/responses.go (update)
package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	apperrors "github.com/bilbothegreedy/server-name-generator/internal/errors"
	"github.com/go-playground/validator/v10"
)

// Common errors
var (
	ErrRequestTimeout = errors.New("request timed out")
)

// ContextKey type for context values
type ContextKey string

// Context keys
const (
	RequestIDKey ContextKey = "request_id"
)

// CustomResponseWriter extends http.ResponseWriter to track status code
type CustomResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// NewResponseWriter creates a new CustomResponseWriter
func NewResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code and passes it to the underlying ResponseWriter
func (rw *CustomResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ErrorResponse represents an error response to be sent to the client
type ErrorResponse struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Detail    string      `json:"detail,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
	Code      string      `json:"code,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// SuccessResponse represents a success response to be sent to the client
type SuccessResponse struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

// RespondWithError sends an error response to the client
// Keep the original function for backward compatibility
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, ErrorResponse{
		Status:  code,
		Message: message,
	})
}

// RespondWithAppError sends an application error response to the client
func RespondWithAppError(w http.ResponseWriter, ctx context.Context, err error) {
	var response ErrorResponse
	var statusCode int

	// Check if it's an application error
	if appErr, ok := err.(*apperrors.AppError); ok {
		statusCode = appErr.StatusCode
		response = ErrorResponse{
			Status:    appErr.StatusCode,
			Message:   appErr.Message,
			Detail:    appErr.Detail,
			RequestID: appErr.RequestID,
			Code:      appErr.Code,
		}
	} else {
		// Generic error
		statusCode = http.StatusInternalServerError
		response = ErrorResponse{
			Status:  statusCode,
			Message: "An unexpected error occurred",
		}

		// Add request ID if present in context
		if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
			response.RequestID = requestID
		}
	}

	// Send the response
	RespondWithJSON(w, statusCode, response)
}

// RespondWithJSON sends a JSON response to the client
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":500,"message":"Failed to marshal JSON response"}`))
		return
	}

	// Clear any previous content
	w.Header().Del("Content-Type")

	// Set headers correctly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	// Write the response body
	w.Write(response)
}

// CreateReadCloser creates an io.ReadCloser from a struct for re-reading request bodies
func CreateReadCloser(v interface{}) io.ReadCloser {
	data, _ := json.Marshal(v)
	return io.NopCloser(bytes.NewReader(data))
}

// ContextWithTimeout creates a context with a timeout
func ContextWithTimeout(ctx context.Context, seconds int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
}

// Validate validates a struct using the validator package
func Validate(s interface{}) error {
	validate := validator.New()

	// Register function to get json tag as field name
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err := validate.Struct(s)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			messages := make([]string, 0, len(validationErrors))
			for _, e := range validationErrors {
				// Get the field name from the json tag
				field := e.Field()

				// Create a user-friendly error message
				var message string
				switch e.Tag() {
				case "required":
					message = fmt.Sprintf("%s is required", field)
				case "min":
					message = fmt.Sprintf("%s must be at least %s characters long", field, e.Param())
				case "max":
					message = fmt.Sprintf("%s must be at most %s characters long", field, e.Param())
				case "uuid":
					message = fmt.Sprintf("%s must be a valid UUID", field)
				default:
					message = fmt.Sprintf("%s failed validation: %s", field, e.Tag())
				}

				messages = append(messages, message)
			}
			return errors.New(strings.Join(messages, "; "))
		}
		return err
	}
	return nil
}
