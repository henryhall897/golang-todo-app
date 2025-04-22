package domain

import "time"

const (
	DefaultLimit     = 10
	DefaultOffset    = 0
	RedisPrefix      = "user"
	RedisEmailPrefix = "email"
	RedisTTL         = 10 * time.Minute
)
