package redis

import (
	"context"
	"time"
)

//go:generate moq -out=../../gen/mocks/redismock/redis_mock.go -pkg=redismock . Cache
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	SetPointer(ctx context.Context, key string, targetKey string, ttl time.Duration) error
	GetPointer(ctx context.Context, key string) (string, error)
}
