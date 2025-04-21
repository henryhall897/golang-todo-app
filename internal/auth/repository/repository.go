package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/henryhall897/golang-todo-app/gen/queries/authstore"
	"github.com/henryhall897/golang-todo-app/internal/auth/domain"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	pool  *pgxpool.Pool
	query *authstore.Queries
}

func New(pool *pgxpool.Pool) *repository {
	return &repository{
		pool:  pool,
		query: authstore.New(pool),
	}
}
func (r *repository) CreateAuthIdentity(ctx context.Context, input domain.CreateAuthIdentityParams) (domain.AuthIdentity, error) {
	// Convert the CreateAuthIdentityParams to InsertAuthIdentityParams
	pgInput := createAuthIdentityParamsToPG(input)

	// Execute the insert query
	created, err := r.query.CreateAuthIdentity(ctx, pgInput)
	if err != nil {
		// Check if the error is a unique constraint violation (e.g., duplicate auth_id)
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // unique_violation
			if pgErr.ConstraintName == "auth_identities_pkey" {
				return domain.AuthIdentity{}, fmt.Errorf("%w", ErrAuthIDAlreadyExists)
			}
			return domain.AuthIdentity{}, fmt.Errorf("unique constraint violation: %w", err)
		}
		return domain.AuthIdentity{}, fmt.Errorf("failed to create auth identity: %w", err)
	}

	result, err := pgToAuthIdentity(created)
	if err != nil {
		return domain.AuthIdentity{}, err
	}

	return result, nil
}

func (r *repository) GetAuthIdentityByAuthID(ctx context.Context, authID string) (domain.AuthIdentity, error) {
	// Execute the query to get the auth identity by AuthID
	identity, err := r.query.GetAuthIdentityByAuthID(ctx, authID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AuthIdentity{}, fmt.Errorf("auth identity %s: %w", authID, common.ErrNotFound)
	} else if err != nil {
		return domain.AuthIdentity{}, fmt.Errorf("auth identity %s: %w", authID, common.ErrInternalServerError)
	}

	// Convert the raw database results into the domain.AuthIdentity type
	result, err := pgToAuthIdentity(identity)
	if err != nil {
		return domain.AuthIdentity{}, err
	}

	return result, nil
}
