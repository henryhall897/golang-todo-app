package users

import (
	"testing"
	"time"

	"github.com/henryhall897/golang-todo-app/internal/users/gen"

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

func TestDBToUsers(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	validTime := time.Now().UTC()

	genUser := gen.User{
		ID: pgtype.UUID{
			Bytes: validUUID,
			Valid: true,
		},
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		CreatedAt: pgtype.Timestamp{Time: validTime, Valid: true}, // Use Timestamptz
		UpdatedAt: pgtype.Timestamp{Time: validTime, Valid: true}, // Use Timestamptz
	}

	// Act
	user, err := dbToUsers(genUser) // Direct call

	// Assert
	require.NoError(t, err)
	require.Equal(t, validUUID, user.ID)
	require.Equal(t, genUser.Name, user.Name)
	require.Equal(t, genUser.Email, user.Email)
	require.Equal(t, &validTime, user.CreatedAt)
	require.Equal(t, validTime, *user.UpdatedAt)
}

func TestDBToUsersInvalidUUID(t *testing.T) {
	// Arrange
	genUser := gen.User{
		ID: pgtype.UUID{
			Valid: false, // Invalid UUID
		},
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	// Act
	_, err := dbToUsers(genUser) // Direct call

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid user id")
}

func TestDBToUsersInvalidTimestamp(t *testing.T) {
	// Arrange
	validUUID := uuid.New()

	genUser := gen.User{
		ID: pgtype.UUID{
			Bytes: validUUID,
			Valid: true,
		},
		Name:      "Jane Doe",
		Email:     "jane.doe@example.com",
		CreatedAt: pgtype.Timestamp{Valid: false}, // Invalid timestamp
		UpdatedAt: pgtype.Timestamp{Valid: false}, // Invalid timestamp
	}

	// Act
	user, err := dbToUsers(genUser) // Direct call

	// Assert
	require.NoError(t, err)
	require.Equal(t, validUUID, user.ID)
	require.Equal(t, genUser.Name, user.Name)
	require.Equal(t, genUser.Email, user.Email)
	require.Nil(t, user.CreatedAt, "Expected CreatedAt to be nil")
	require.Nil(t, user.UpdatedAt, "Expected UpdatedAt to be nil")
}

func TestToDBUpdateUserParams(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	name := "Jane Doe"
	email := "jane.doe@example.com"

	inputParams := UpdateUserParams{
		ID:    validUUID,
		Name:  name,
		Email: email,
	}

	// Act
	dbParams, err := toDBUpdateUserParams(inputParams) // Direct call

	// Assert
	require.NoError(t, err)
	require.Equal(t, validUUID, uuid.UUID(dbParams.ID.Bytes), "UUID should match")
	require.Equal(t, name, dbParams.Name, "Name should match")
	require.Equal(t, email, dbParams.Email, "Email should match")
}

func TestToDBCreateUserParams(t *testing.T) {
	// Arrange
	name := "John Doe"
	email := "john.doe@example.com"

	inputParams := CreateUserParams{
		Name:  name,
		Email: email,
	}

	// Act
	dbParams := toDBCreateUserParams(inputParams) // Direct call

	// Assert
	require.Equal(t, name, dbParams.Name, "Name should match")
	require.Equal(t, email, dbParams.Email, "Email should match")
}
