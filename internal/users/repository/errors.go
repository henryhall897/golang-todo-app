package repository

import (
	"errors"
)

// Errors specific to the repository
var (
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidDbUserID     = errors.New("invalid user id")
	ErrFailedToParseUUID   = errors.New("failed to parse uuid")
	ErrAuthIDAlreadyExists = errors.New("auth id already exists")
)
