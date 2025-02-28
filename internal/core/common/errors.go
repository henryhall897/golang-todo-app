package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrNotFound indicates that a record queried for in the database was not found.
var ErrNotFound = errors.New("not found")
var ErrValidation = errors.New("validation failed")
var ErrInternalServerError = errors.New("internal server error")
var ErrInvalidUUID = errors.New("invalid UUID")

// ErrorResponse represents a standardized JSON error response.
type ErrorResponse struct {
	Code    int    `json:"code"`    // HTTP status code
	Message string `json:"message"` // Error message for the client
}

// WriteJSONError sends a standardized JSON error response.
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(ErrorResponse{
		Code:    statusCode,
		Message: message,
	})
	if err != nil {
		// Log error instead of sending another response (prevents double-write)
		fmt.Printf("Failed to encode JSON error response: %v\n", err)
	}
}
