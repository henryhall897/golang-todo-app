package cache

import (
	"context"
	"fmt"
	"time"

	redispkg "github.com/henryhall897/golang-todo-app/pkg/redis"
)

const blacklistKeyPrefix = "blacklist"

// RedisAdapter implements the auth.Cache interface
type RedisAdapter struct {
	client redispkg.Cache
}

// NewRedisAdapter creates a new instance of Redis-backed auth cache
func NewRedisAdapter(client redispkg.Cache) *RedisAdapter {
	return &RedisAdapter{client: client}
}

func (r *RedisAdapter) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	key := r.buildKey(jti)
	return r.client.Set(ctx, key, true, ttl)
}

func (r *RedisAdapter) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := r.buildKey(jti)
	var blacklisted bool
	err := r.client.Get(ctx, key, &blacklisted)
	if err != nil {
		// Treat missing key as not blacklisted
		return false, nil
	}
	return blacklisted, nil
}

func (r *RedisAdapter) buildKey(jti string) string {
	return fmt.Sprintf("%s:%s", blacklistKeyPrefix, jti)
}
