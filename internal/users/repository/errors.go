package repository

import (
	"errors"
)

// Errors specific to the repository
var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrInvalidDbUserID = errors.New("invalid user id")
var ErrFailedToParseUUID = errors.New("failed to parse uuid")
