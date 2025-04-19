package cache

import (
	"context"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
	redispkg "github.com/henryhall897/golang-todo-app/pkg/redis"
)

type RedisUser struct {
	genericCache redispkg.Cache // this is your generic cache interface (Get, Set, Delete)
}

func NewRedisUser(generic redispkg.Cache) *RedisUser {
	return &RedisUser{genericCache: generic}
}

// setting cache functions
// CacheUserByID caches a user by ID
func (c *RedisUser) CacheUserByID(ctx context.Context, user domain.User) error {
	key := domain.CacheKeyByID(user.ID)
	return c.genericCache.Set(ctx, key, user, domain.RedisTTL)
}

// CacheUserByEmail caches a user by email
func (c *RedisUser) CacheUserByEmail(ctx context.Context, user domain.User) error {
	key := domain.CacheKeyByEmail(user.Email)
	return c.genericCache.Set(ctx, key, user, domain.RedisTTL)
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

// GetUserByEmail retrieves a user by email from the cache
func (c *RedisUser) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	key := domain.CacheKeyByEmail(email)
	var user domain.User
	err := c.genericCache.Get(ctx, key, &user)
	return user, err
}

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
