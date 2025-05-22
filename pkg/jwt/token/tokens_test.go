package token

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/pkg/jwt/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTTokenGenerator_GenerateAndParse(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	role := "user"

	gen := NewJWTTokenGenerator(TokenConfig{
		SecretKey:     "test-secret",
		TokenDuration: time.Hour,
		Issuer:        "jwt-test",
	})

	t.Run("successfully generates and parses a valid token", func(t *testing.T) {
		// Construct fake user
		user := domain.Payload{
			ID:   userID,
			Role: role,
		}

		// Generate token
		tokenStr, err := gen.Gen(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenStr)

		// Parse token
		claims, err := gen.Parse(ctx, tokenStr)
		require.NoError(t, err)

		// Verify claims
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, "jwt-test", claims.Issuer)
		assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, time.Second)
		assert.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, 5*time.Second)
		assert.NotEmpty(t, claims.ID)
	})

	t.Run("fails to parse a malformed token", func(t *testing.T) {
		_, err := gen.Parse(ctx, "this.is.not.a.valid.token")
		require.Error(t, err)
	})

	t.Run("fails on signature mismatch", func(t *testing.T) {
		// Create a token with a different secret
		altGen := NewJWTTokenGenerator(TokenConfig{
			SecretKey:     "wrong-secret",
			TokenDuration: time.Hour,
			Issuer:        "jwt-test",
		})

		tokenStr, err := altGen.Gen(ctx, domain.Payload{
			ID:   userID,
			Role: role,
		})
		require.NoError(t, err)

		// Parse with original generator → should fail
		_, err = gen.Parse(ctx, tokenStr)
		require.Error(t, err)
	})
}
