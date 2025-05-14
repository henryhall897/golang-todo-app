package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/gen/queries/authstore"
	"github.com/henryhall897/golang-todo-app/internal/auth/testutils"
	usertestutils "github.com/henryhall897/golang-todo-app/internal/users/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"
)

type authTransformTestSuite struct {
	suite.Suite
}

func (suite *authTransformTestSuite) SetupSuite() {
	// No setup needed
}

func TestAuthTransform(t *testing.T) {
	suite.Run(t, new(authTransformTestSuite))
}

func (suite *authTransformTestSuite) TestPGToAuthIdentity() {
	genAuth := testutils.GenerateMockPGAuthIdentity()

	suite.T().Run("Valid Auth Identity", func(t *testing.T) {
		auth, err := pgToAuthIdentity(genAuth)
		require.NoError(t, err)
		require.Equal(t, genAuth.AuthID, auth.AuthID)
		require.Equal(t, genAuth.Provider, auth.Provider)
		require.Equal(t, uuid.UUID(genAuth.UserID.Bytes), auth.UserID)
		require.Equal(t, genAuth.CreatedAt.Time, auth.CreatedAt)
	})

	suite.T().Run("Invalid Timestamp", func(t *testing.T) {
		invalidAuth := genAuth
		invalidAuth.CreatedAt.Valid = false

		auth, err := pgToAuthIdentity(invalidAuth)
		require.NoError(t, err)
		require.Equal(t, time.Time{}, auth.CreatedAt)
		require.Equal(t, time.Time{}, auth.UpdatedAt)
	})
}

func (suite *authTransformTestSuite) TestCreateAuthIdentityParamsToPG() {
	suite.T().Run("Valid Input", func(t *testing.T) {
		// Arrange
		mockUsers := usertestutils.GenerateMockUsers(1)
		mockAuthParams := testutils.GenerateMockAuthParams(mockUsers)
		domainParams := mockAuthParams[0]

		// Act
		pgParams := createAuthIdentityParamsToPG(domainParams)

		// Assert
		require.Equal(t, domainParams.AuthID, pgParams.AuthID)
		require.Equal(t, domainParams.Provider, pgParams.Provider)
		require.True(t, pgParams.UserID.Valid)
		require.Equal(t, domainParams.UserID, uuid.UUID(pgParams.UserID.Bytes))
	})
}

func (suite *authTransformTestSuite) TestPGToAuthIdentitiesSlice() {
	genAuth := testutils.GenerateMockPGAuthIdentity()

	suite.T().Run("Valid Auth Identity Slice", func(t *testing.T) {
		input := []authstore.AuthIdentity{genAuth, genAuth} // test multiple
		result, err := pgToAuthIdentitiesSlice(input)

		require.NoError(t, err)
		require.Len(t, result, len(input))

		for i := range result {
			require.Equal(t, genAuth.AuthID, result[i].AuthID)
			require.Equal(t, genAuth.Provider, result[i].Provider)
			require.Equal(t, uuid.UUID(genAuth.UserID.Bytes), result[i].UserID)
			require.Equal(t, genAuth.CreatedAt.Time, result[i].CreatedAt)
		}
	})

	suite.T().Run("One Invalid Timestamp", func(t *testing.T) {
		invalid := genAuth
		invalid.CreatedAt.Valid = false

		input := []authstore.AuthIdentity{genAuth, invalid}
		result, err := pgToAuthIdentitiesSlice(input)

		require.NoError(t, err)
		require.Len(t, result, 2)

		assert.Equal(t, genAuth.CreatedAt.Time, result[0].CreatedAt)
		assert.Equal(t, time.Time{}, result[1].CreatedAt)
		assert.Equal(t, time.Time{}, result[1].UpdatedAt)
	})
}
