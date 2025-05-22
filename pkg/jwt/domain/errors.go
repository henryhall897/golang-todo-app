package domain

import "errors"

var (
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenInvalid     = errors.New("token is invalid")
	ErrTokenBlacklisted = errors.New("token is blacklisted")
)
