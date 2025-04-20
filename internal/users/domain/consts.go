package domain

import "time"

const (
	DefaultLimit      = 10
	DefaultOffset     = 0
	RedisPrefix       = "user"
	RedisEmailPrefix  = "email"
	RedisAuthIDPrefix = "auth_id"
	RedisTTL          = 10 * time.Minute
)
