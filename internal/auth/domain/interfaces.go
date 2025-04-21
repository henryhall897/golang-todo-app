package domain

import "context"

type Repository interface {
	// CreateAuthIdentity creates a new auth identity in the database.
	CreateAuthIdentity(ctx context.Context, input CreateAuthIdentityParams) (AuthIdentity, error)
	// GetAuthIdentityByAuthID retrieves an auth identity by its auth ID.
	GetAuthIdentityByAuthID(ctx context.Context, authID string) (AuthIdentity, error)
}
