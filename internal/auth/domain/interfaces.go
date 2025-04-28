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
	GetAuthIdentityByUserID(ctx context.Context, userID uuid.UUID) (AuthIdentity, error)
	UpdateAuthIdentityRole(ctx context.Context, params UpdateAuthIdentityParams) (AuthIdentity, error)
	DeleteAuthIdentityByAuthID(ctx context.Context, authID string) error
}
