package common

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrNotFound indicates that a record queried for in the database was not found.
var ErrNotFound = errors.New("not found")
var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrInvalidRequestBody = errors.New("invalid request body")

// ErrorResponse represents a standardized JSON error response.
type ErrorResponse struct {
	Code    int    `json:"code"`    // HTTP status code
	Message string `json:"message"` // Error message for the client
}

// WriteJSONError sends a standardized JSON error response.
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    statusCode,
		Message: message,
	})
}
