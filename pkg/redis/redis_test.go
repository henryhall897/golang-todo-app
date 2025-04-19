package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type RedisTestSuite struct {
	Cache  *JSONCache
	Server *miniredis.Miniredis
	logger *zap.SugaredLogger
}

type TestUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// TestUser is a struct for testing purposes.
func SetupSuite() *RedisTestSuite {
	srv, err := miniredis.Run()
	if err != nil {
		panic(fmt.Sprintf("Failed to start miniredis: %v", err))
	}

	client := redis.NewClient(&redis.Options{
		Addr: srv.Addr(),
	})

	logger := zap.NewNop().Sugar()
	cache := NewJSONCache(client, "test", logger)

	return &RedisTestSuite{
		Cache:  cache,
		Server: srv,
		logger: logger,
	}
}

func generateTestUsers(count int) []TestUser {
	users := make([]TestUser, count)
	for i := 0; i < count; i++ {
		users[i] = TestUser{
			ID:    fmt.Sprintf("user-%d", i+1),
			Email: fmt.Sprintf("user%d@example.com", i+1),
		}
	}
	return users
}

func TestJSONCache_Behavior(t *testing.T) {
	suite := SetupSuite()
	defer suite.Server.Close()

	ctx := context.Background()

	t.Run("Set and Get - success", func(t *testing.T) {
		users := generateTestUsers(1)
		user := users[0]
		key := user.ID

		err := suite.Cache.Set(ctx, key, user, time.Minute)
		require.NoError(t, err)

		storedJSON, err := suite.Server.Get("test:" + key)
		require.NoError(t, err)

		var parsed TestUser
		require.NoError(t, json.Unmarshal([]byte(storedJSON), &parsed))
		assert.Equal(t, user, parsed)

		var result TestUser
		err = suite.Cache.Get(ctx, key, &result)
		require.NoError(t, err)
		assert.Equal(t, user, result)
	})

	t.Run("Cache miss - key does not exist", func(t *testing.T) {
		key := "nonexistent-user"
		var result TestUser

		err := suite.Cache.Get(ctx, key, &result)
		require.ErrorIs(t, err, redis.Nil, "Expected redis.Nil on cache miss")
	})

	t.Run("Delete - success", func(t *testing.T) {
		users := generateTestUsers(1)
		user := users[0]
		key := user.ID

		// Set key first
		err := suite.Cache.Set(ctx, key, user, time.Minute)
		require.NoError(t, err)

		fullKey := "test:" + key
		require.True(t, suite.Server.Exists(fullKey), "Key should exist before deletion")

		// Delete the key
		err = suite.Cache.Delete(ctx, key)
		require.NoError(t, err)

		// Confirm deletion
		assert.False(t, suite.Server.Exists(fullKey), "Key should be removed after deletion")
	})
}
