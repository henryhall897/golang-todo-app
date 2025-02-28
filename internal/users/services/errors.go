package services

import "errors"

// Service-level errors (handler should only see these)
var (
	ErrEmailAlreadyExists = errors.New("email already exists")
)
