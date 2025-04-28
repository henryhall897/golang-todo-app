package repository

import (
	"fmt"

	"github.com/henryhall897/golang-todo-app/gen/queries/authstore"
	"github.com/henryhall897/golang-todo-app/internal/auth/domain"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/jackc/pgx/v5/pgtype"
)

// pgToAuthIdentities converts a slice of authstore.AuthIdentity to a slice of domain.AuthIdentity
func pgToAuthIdentity(pg authstore.AuthIdentity) (domain.AuthIdentity, error) {
	userID, err := common.FromPgUUID(pg.UserID)
	if err != nil {
		return domain.AuthIdentity{}, fmt.Errorf("failed to transform user_id uuid: %w", err)
	}

	createdAt := common.FromPgTimestamp(pg.CreatedAt)
	updatedAt := common.FromPgTimestamp(pg.UpdatedAt)

	return domain.AuthIdentity{
		AuthID:    pg.AuthID,
		Provider:  pg.Provider,
		UserID:    userID,
		Role:      pg.Role,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// pgToAuthIdentities converts a slice of authstore.AuthIdentity to a slice of domain.AuthIdentity
func createAuthIdentityParamsToPG(params domain.CreateAuthIdentityParams) authstore.CreateAuthIdentityParams {
	return authstore.CreateAuthIdentityParams{
		AuthID:   params.AuthID,
		Provider: params.Provider,
		UserID:   pgtype.UUID{Bytes: params.UserID, Valid: true},
		Role:     params.Role,
	}
}

// updateAuthIdentityParamsToPG converts a domain.UpdateAuthIdentityParams to authstore.UpdateAuthIdentityParams
func updateAuthIdentityParamsToPG(params domain.UpdateAuthIdentityParams) authstore.UpdateAuthIdentityRoleParams {
	return authstore.UpdateAuthIdentityRoleParams{
		AuthID: params.AuthID,
		Role:   params.Role,
	}
}
