package users

import (
	"testing"
	"time"

	"golang-todo-app/internal/users/gen"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TransformTestSuite struct {
	suite.Suite
}

func TestTransform(t *testing.T) {
	suite.Run(t, new(TransformTestSuite))
}

func TestPgToUsers(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	validTime := time.Now()

	genUser := gen.User{
		ID: pgtype.UUID{
			Bytes: validUUID,
			Valid: true,
		},
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		CreatedAt: pgtype.Timestamptz{Time: validTime, Valid: true}, // Use Timestamptz
		UpdatedAt: pgtype.Timestamptz{Time: validTime, Valid: true}, // Use Timestamptz
	}

	// Act
	user, err := pgToUsers(genUser) // Direct call

	// Assert
	require.NoError(t, err)
	require.Equal(t, validUUID, user.ID)
	require.Equal(t, genUser.Name, user.Name)
	require.Equal(t, genUser.Email, user.Email)
	require.Equal(t, validTime, user.CreatedAt)
	require.Equal(t, validTime, user.UpdatedAt)
}

func TestPgToUsersInvalidUUID(t *testing.T) {
	// Arrange
	genUser := gen.User{
		ID: pgtype.UUID{
			Valid: false, // Invalid UUID
		},
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	// Act
	_, err := pgToUsers(genUser) // Direct call

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid user id")
}

func TestPgToUsersInvalidTimestamp(t *testing.T) {
	// Arrange
	validUUID := uuid.New()

	genUser := gen.User{
		ID: pgtype.UUID{
			Bytes: validUUID,
			Valid: true,
		},
		Name:      "Jane Doe",
		Email:     "jane.doe@example.com",
		CreatedAt: pgtype.Timestamptz{Valid: false}, // Invalid timestamp
		UpdatedAt: pgtype.Timestamptz{Valid: false}, // Invalid timestamp
	}

	// Act
	user, err := pgToUsers(genUser) // Direct call

	// Assert
	require.NoError(t, err) // Expect no error as pgToUsers doesn't validate timestamps
	require.Equal(t, validUUID, user.ID)
	require.Equal(t, genUser.Name, user.Name)
	require.Equal(t, genUser.Email, user.Email)
	require.Equal(t, time.Time{}, user.CreatedAt) // Expect zero value for invalid timestamp
	require.Equal(t, time.Time{}, user.UpdatedAt) // Expect zero value for invalid timestamp
}

func TestUUIDToPgUUID(t *testing.T) {
	// Test valid UUID conversion
	validUUID := uuid.New()
	pgUUID, err := uuidToPgUUID(validUUID) // Direct call
	require.NoError(t, err)
	require.True(t, pgUUID.Valid)

	// Convert pgUUID.Bytes back to uuid.UUID for comparison
	convertedUUID, err := uuid.FromBytes(pgUUID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validUUID, convertedUUID)

	// Test nil UUID
	nilUUID := uuid.Nil
	_, err = uuidToPgUUID(nilUUID) // Direct call
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid UUID")
}
