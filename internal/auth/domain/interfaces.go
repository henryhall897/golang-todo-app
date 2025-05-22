package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	udomain "github.com/henryhall897/golang-todo-app/internal/users/domain"
)

//go:generate moq -out=../../../gen/mocks/authmocks/repository_mock.go -pkg=authmocks . Repository
type Repository interface {
	// CreateAuthIdentity creates a new auth identity in the database.
	CreateAuthIdentity(ctx context.Context, input CreateAuthIdentityParams) (AuthIdentity, error)
	// GetAuthIdentityByAuthID retrieves an auth identity by its auth ID.
	GetAuthIdentityByAuthID(ctx context.Context, authID string) (AuthIdentity, error)
	GetAuthIdentitiesByUserID(ctx context.Context, userID uuid.UUID) ([]AuthIdentity, error)
	DeleteAuthIdentityByAuthID(ctx context.Context, authID string) error
}

//go:generate moq -out=../../../gen/mocks/authmocks/cache_mock.go -pkg=authmocks . Cache
type Cache interface {
	// SetAuthIdentity sets the auth identity in the cache.
	SetAuthIdentity(ctx context.Context, authIdentity AuthIdentity) error
	// GetAuthIdentity retrieves the auth identity from the cache.
	GetAuthIdentity(ctx context.Context, authID string) (AuthIdentity, error)
	// DeleteAuthIdentity deletes the auth identity from the cache.
	DeleteAuthIdentity(ctx context.Context, authID string) error
	BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}

//go:generate moq -out=../../../gen/mocks/authmocks/service_mock.go -pkg=authmocks . Service
type Service interface {
	// LoginOrRegister logs in or registers a user based on the provided auth ID and provider.
	LoginOrRegister(ctx context.Context, input AuthLoginParams) (string, udomain.User, error)
	// Logout logs out a user by invalidating their token.
	Logout(ctx context.Context, token string) error
}
