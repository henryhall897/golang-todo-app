package repository

import (
	"testing"
	"time"

	"github.com/henryhall897/golang-todo-app/gen/queries/userstore"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type transformTestSuite struct {
	suite.Suite
}

func (suite *transformTestSuite) SetupSuite() {
}

func TestTransform(t *testing.T) {
	suite.Run(t, new(transformTestSuite))
}

func (suite *transformTestSuite) TestPGToUsers() {
	// Arrange: Create a base valid user
	validUUID := uuid.New()
	validTime := time.Now().UTC()

	genUser := userstore.User{
		ID:        pgtype.UUID{Bytes: validUUID, Valid: true},
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		CreatedAt: pgtype.Timestamp{Time: validTime, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: validTime, Valid: true},
	}

	// Subtest: Valid User
	suite.T().Run("Valid User", func(t *testing.T) {
		user, err := pgToUsers(genUser)

		require.NoError(t, err)
		require.Equal(t, uuid.UUID(genUser.ID.Bytes), user.ID)
		require.Equal(t, genUser.Name, user.Name)
		require.Equal(t, genUser.Email, user.Email)
		require.NotNil(t, user.CreatedAt)
		require.NotNil(t, user.UpdatedAt)
		require.Equal(t, genUser.CreatedAt.Time, user.CreatedAt)
		require.Equal(t, genUser.UpdatedAt.Time, user.UpdatedAt)
	})

	// Subtest: Invalid Timestamp
	suite.T().Run("Invalid Timestamp", func(t *testing.T) {
		invalidTimestampUser := genUser
		invalidTimestampUser.CreatedAt = pgtype.Timestamp{Valid: false}
		invalidTimestampUser.UpdatedAt = pgtype.Timestamp{Valid: false}

		user, err := pgToUsers(invalidTimestampUser)
		require.NoError(t, err)
		require.Equal(t, uuid.UUID(invalidTimestampUser.ID.Bytes), user.ID)
		require.Equal(t, invalidTimestampUser.Name, user.Name)
		require.Equal(t, invalidTimestampUser.Email, user.Email)

		// Updated: check for zero time instead of nil
		require.Equal(t, time.Time{}, user.CreatedAt, "Expected CreatedAt to be zero time")
		require.Equal(t, time.Time{}, user.UpdatedAt, "Expected UpdatedAt to be zero time")
	})

}

func (suite *transformTestSuite) TestToDBUpdateUserParams() {
	// Arrange
	validUUID := uuid.New()
	name := "Jane Doe"
	email := "jane.doe@example.com"

	inputParams := domain.UpdateUserParams{
		ID:    validUUID,
		Name:  name,
		Email: email,
	}

	// Act
	dbParams, err := updateUserParamsToPG(inputParams)

	// Assert
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), validUUID, uuid.UUID(dbParams.ID.Bytes), "UUID should match")
	require.Equal(suite.T(), name, dbParams.Name, "Name should match")
	require.Equal(suite.T(), email, dbParams.Email, "Email should match")
}

func (suite *transformTestSuite) TestToDBCreateUserParams() {
	// Arrange
	name := "John Doe"
	email := "john.doe@example.com"

	inputParams := domain.CreateUserParams{
		Name:  name,
		Email: email,
	}

	// Act
	dbParams := createUserParamsToPG(inputParams)

	// Assert
	require.Equal(suite.T(), name, dbParams.Name, "Name should match")
	require.Equal(suite.T(), email, dbParams.Email, "Email should match")
}
