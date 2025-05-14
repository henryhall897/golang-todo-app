package domain

import "time"

type Role string

const (

	// Pagination constants
	DefaultLimit  = 10
	DefaultOffset = 0

	// Redis constants
	RedisPrefix      = "user"
	RedisEmailPrefix = "email"
	RedisTTL         = 10 * time.Minute

	// Role constants
	DefaultRole Role = "user"
	RoleAdmin   Role = "admin"
)
