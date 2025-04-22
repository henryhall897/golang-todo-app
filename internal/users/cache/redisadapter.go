package cache

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
	redispkg "github.com/henryhall897/golang-todo-app/pkg/redis"
	"github.com/redis/go-redis/v9"
)

type RedisUser struct {
	genericCache redispkg.Cache // this is your generic cache interface (Get, Set, Delete)
}

func NewRedisUser(generic redispkg.Cache) *RedisUser {
	return &RedisUser{genericCache: generic}
}

// setting cache functions
// CacheUserByID caches a user by ID
func (c *RedisUser) cacheUserByID(ctx context.Context, user domain.User) error {
	key := domain.CacheKeyByID(user.ID)
	return c.genericCache.Set(ctx, key, user, domain.RedisTTL)
}

// CacheUserByEmail stores a pointer from email → user ID and caches the full user by ID
func (c *RedisUser) cacheEmailPointer(ctx context.Context, user domain.User) error {
	// Create pointer from email → UUID string
	pointerKey := domain.CacheKeyByEmail(user.Email)
	return c.genericCache.SetPointer(ctx, pointerKey, domain.CacheKeyByID(user.ID), domain.RedisTTL)
}

/*// SetAuthIDPointer sets a pointer to a user by Auth0 ID in the cache
func (c *RedisUser) cacheAuthIDPointer(ctx context.Context, user domain.User) error {
	pointerKey := domain.CacheKeyByAuthID(user.AuthID)
	targetKey := domain.CacheKeyByID(user.ID)
	return c.genericCache.SetPointer(ctx, pointerKey, targetKey, domain.RedisTTL)
}*/

// CacheUser caches a user by ID, email, and AuthID
func (c *RedisUser) CacheUser(ctx context.Context, user domain.User) error {
	if err := c.cacheUserByID(ctx, user); err != nil {
		return fmt.Errorf("failed to cache user by ID: %w", err)
	}
	/*if err := c.cacheAuthIDPointer(ctx, user); err != nil {
		return fmt.Errorf("failed to cache all user pointers: %w", err)
	}*/
	if err := c.cacheEmailPointer(ctx, user); err != nil {
		return fmt.Errorf("failed to cache user by email: %w", err)
	}
	return nil
}

// CacheUserByPagination caches a list of users by pagination parameters
func (c *RedisUser) CacheUserByPagination(ctx context.Context, users []domain.User, params domain.GetUsersParams) error {
	key := domain.CacheKeyByPagination(params.Limit, params.Offset)
	return c.genericCache.Set(ctx, key, users, domain.RedisTTL)
}

// Get from Cache Functions
// GetUserByID retrieves a user by ID from the cache
func (c *RedisUser) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	key := domain.CacheKeyByID(id)
	var user domain.User
	err := c.genericCache.Get(ctx, key, &user)
	return user, err
}

// GetUserByEmail resolves the email → user ID pointer and returns the full cached user
func (c *RedisUser) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	pointerKey := domain.CacheKeyByEmail(email)

	idStr, err := c.genericCache.GetPointer(ctx, pointerKey)
	if err != nil {
		return domain.User{}, err
	}
	if idStr == "" {
		return domain.User{}, redis.Nil // cache miss
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		return domain.User{}, fmt.Errorf("invalid UUID in email pointer: %w", err)
	}

	return c.GetUserByID(ctx, userID)
}

/*// GetUserByAuthID resolves the Auth0 ID pointer and returns the full cached user
func (c *RedisUser) GetUserByAuthID(ctx context.Context, authID string) (domain.User, error) {
	pointerKey := domain.CacheKeyByAuthID(authID)

	idStr, err := c.genericCache.GetPointer(ctx, pointerKey)
	if err != nil {
		return domain.User{}, err
	}
	if idStr == "" {
		return domain.User{}, redis.Nil // cache miss
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		return domain.User{}, fmt.Errorf("invalid UUID in auth_id pointer: %w", err)
	}

	return c.GetUserByID(ctx, userID)
}*/

// GetUserByPagination retrieves a list of users by pagination parameters from the cache
func (c *RedisUser) GetUserByPagination(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
	key := domain.CacheKeyByPagination(params.Limit, params.Offset)
	var users []domain.User
	err := c.genericCache.Get(ctx, key, &users)
	return users, err
}

// Delete from Cache Functions
// DeleteUserByID deletes a user by ID from the cache
func (c *RedisUser) DeleteUserByID(ctx context.Context, id uuid.UUID) error {
	key := domain.CacheKeyByID(id)
	return c.genericCache.Delete(ctx, key)
}

// DeleteUserByEmail deletes a user by email from the cache
func (c *RedisUser) DeleteUserByEmail(ctx context.Context, email string) error {
	key := domain.CacheKeyByEmail(email)
	return c.genericCache.Delete(ctx, key)
}
