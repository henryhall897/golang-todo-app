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

	return domain.AuthIdentity{
		AuthID:    pg.AuthID,
		Provider:  pg.Provider,
		UserID:    userID,
		CreatedAt: createdAt,
	}, nil
}

// pgToAuthIdentities converts a slice of authstore.AuthIdentity to a slice of domain.AuthIdentity
func createAuthIdentityParamsToPG(params domain.CreateAuthIdentityParams) authstore.CreateAuthIdentityParams {
	return authstore.CreateAuthIdentityParams{
		AuthID:   params.AuthID,
		Provider: params.Provider,
		UserID:   pgtype.UUID{Bytes: params.UserID, Valid: true},
	}
}

func pgToAuthIdentitiesSlice(auths []authstore.AuthIdentity) ([]domain.AuthIdentity, error) {
	results := make([]domain.AuthIdentity, 0, len(auths))
	for _, a := range auths {
		converted, err := pgToAuthIdentity(a)
		if err != nil {
			return nil, fmt.Errorf("failed to convert auth identity: %w", err)
		}
		results = append(results, converted)
	}
	return results, nil
}
