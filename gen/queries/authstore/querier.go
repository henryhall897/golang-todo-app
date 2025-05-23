// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package authstore

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	CreateAuthIdentity(ctx context.Context, arg CreateAuthIdentityParams) (AuthIdentity, error)
	DeleteAuthIdentityByAuthID(ctx context.Context, authID string) (int64, error)
	GetAuthIdentityByAuthID(ctx context.Context, authID string) (AuthIdentity, error)
	GetAuthIdentityByUserID(ctx context.Context, userID pgtype.UUID) (AuthIdentity, error)
	UpdateAuthIdentityRole(ctx context.Context, arg UpdateAuthIdentityRoleParams) (AuthIdentity, error)
}

var _ Querier = (*Queries)(nil)
