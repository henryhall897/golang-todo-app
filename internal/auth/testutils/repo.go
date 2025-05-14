package testutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/gen/queries/authstore"
	domain "github.com/henryhall897/golang-todo-app/internal/auth/domain"
	userdomain "github.com/henryhall897/golang-todo-app/internal/users/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// Repo level
// InsertMockUsersIntoDB inserts a list of mock users into the test DB.
func InsertMockUsersIntoDB(t *testing.T, db *pgxpool.Pool, ctx context.Context, users []userdomain.User) {
	t.Helper()

	for _, user := range users {
		_, err := db.Exec(ctx,
			`INSERT INTO users (id, name, email, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
			user.ID, user.Name, user.Email, user.CreatedAt, user.UpdatedAt,
		)
		require.NoError(t, err, "failed to insert mock user: %s", user.Email)
	}
}

// Transformations
// GenerateMockAuthParams creates mock CreateAuthIdentityParams for a given list of users.
func GenerateMockAuthParams(users []userdomain.User) []domain.CreateAuthIdentityParams {
	authParams := make([]domain.CreateAuthIdentityParams, len(users))

	for i, user := range users {
		authParams[i] = domain.CreateAuthIdentityParams{
			AuthID:   fmt.Sprintf("%s|mock%d", MockProviderBase, i+1),
			Provider: MockProviderBase,
			UserID:   user.ID,
		}
	}

	return authParams
}

// GenerateMockPGAuthIdentity creates a mock authstore.AuthIdentity with valid fields.
func GenerateMockPGAuthIdentity() authstore.AuthIdentity {
	validUUID := uuid.New()
	validTime := time.Now().UTC()

	return authstore.AuthIdentity{
		AuthID:    "auth0|mock123",
		Provider:  "auth0",
		UserID:    pgtype.UUID{Bytes: validUUID, Valid: true},
		CreatedAt: pgtype.Timestamp{Time: validTime, Valid: true},
	}
}
