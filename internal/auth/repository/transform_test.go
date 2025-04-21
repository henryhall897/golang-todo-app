package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/auth/testutils"
	usertestutils "github.com/henryhall897/golang-todo-app/internal/users/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
		require.Equal(t, genAuth.Role, auth.Role)
		require.Equal(t, genAuth.CreatedAt.Time, auth.CreatedAt)
		require.Equal(t, genAuth.UpdatedAt.Time, auth.UpdatedAt)
	})

	suite.T().Run("Invalid Timestamp", func(t *testing.T) {
		invalidAuth := genAuth
		invalidAuth.CreatedAt.Valid = false
		invalidAuth.UpdatedAt.Valid = false

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
		require.Equal(t, domainParams.Role, pgParams.Role)
		require.True(t, pgParams.UserID.Valid)
		require.Equal(t, domainParams.UserID, uuid.UUID(pgParams.UserID.Bytes))
	})
}
