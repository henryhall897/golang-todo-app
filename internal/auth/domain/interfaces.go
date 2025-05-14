package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// CreateAuthIdentity creates a new auth identity in the database.
	CreateAuthIdentity(ctx context.Context, input CreateAuthIdentityParams) (AuthIdentity, error)
	// GetAuthIdentityByAuthID retrieves an auth identity by its auth ID.
	GetAuthIdentityByAuthID(ctx context.Context, authID string) (AuthIdentity, error)
	GetAuthIdentitiesByUserID(ctx context.Context, userID uuid.UUID) ([]AuthIdentity, error)
	DeleteAuthIdentityByAuthID(ctx context.Context, authID string) error
}

type Cache interface {
	// SetAuthIdentity sets the auth identity in the cache.
	SetAuthIdentity(ctx context.Context, authIdentity AuthIdentity) error
	// GetAuthIdentity retrieves the auth identity from the cache.
	GetAuthIdentity(ctx context.Context, authID string) (AuthIdentity, error)
	// DeleteAuthIdentity deletes the auth identity from the cache.
	DeleteAuthIdentity(ctx context.Context, authID string) error
}
