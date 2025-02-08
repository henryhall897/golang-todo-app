package common

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ErrNotFound this indicates that a record queried for in the database was not found
var ErrNotFound = errors.New("not found")

// UserNotFoundError represents an error when a user is not found.
type UserIDNotFoundError struct {
	UserID uuid.UUID
}

// Error implements the error interface for UserNotFoundError.
func (e *UserIDNotFoundError) Error() string {
	return fmt.Sprintf("user with ID %s not found", e.UserID)
}
