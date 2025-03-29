package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var _ Cache = (*JSONCache)(nil) // compile-time interface check

type JSONCache struct {
	client *redis.Client
	prefix string
	logger *zap.SugaredLogger
}

func NewJSONCache(client *redis.Client, prefix string, logger *zap.SugaredLogger) *JSONCache {
	return &JSONCache{
		client: client,
		prefix: prefix,
		logger: logger,
	}
}

func (c *JSONCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Errorw("Failed to serialize data", "key", key, "error", err)
		return err
	}

	namespacedKey := c.prefix + ":" + key
	if err := c.client.Set(ctx, namespacedKey, data, ttl).Err(); err != nil {
		c.logger.Errorw("Failed to set data in Redis", "key", namespacedKey, "error", err)
		return err
	}

	c.logger.Debugw("Data cached", "key", namespacedKey)
	return nil
}

func (c *JSONCache) Get(ctx context.Context, key string, dest interface{}) error {
	namespacedKey := c.prefix + ":" + key
	data, err := c.client.Get(ctx, namespacedKey).Result()
	if err != nil {
		if err == redis.Nil {
			return err // cache miss is expected
		}
		c.logger.Warnw("Failed to get data from Redis", "key", namespacedKey, "error", err)
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		c.logger.Warnw("Failed to deserialize cached data", "key", namespacedKey, "error", err)
		return err
	}

	return nil
}

func (c *JSONCache) Delete(ctx context.Context, key string) error {
	namespacedKey := c.prefix + ":" + key
	if err := c.client.Del(ctx, namespacedKey).Err(); err != nil {
		c.logger.Warnw("Failed to delete key from Redis", "key", namespacedKey, "error", err)
		return err
	}
	return nil
}
