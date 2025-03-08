package common

import (
	"errors"
)

// ErrNotFound indicates that a record queried for in the database was not found.
var ErrNotFound = errors.New("not found")
var ErrValidation = errors.New("validation failed")
var ErrInternalServerError = errors.New("internal server error")
var ErrInvalidUUID = errors.New("invalid UUID")
