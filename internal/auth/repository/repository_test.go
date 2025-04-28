//go:build unit

package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/henryhall897/golang-todo-app/internal/auth/domain"
	"github.com/henryhall897/golang-todo-app/internal/auth/testutils"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/core/dbpool"
	usertestutils "github.com/henryhall897/golang-todo-app/internal/users/testutils"
	"github.com/henryhall897/golang-todo-app/pkg/dbtest"
)

type AuthTestSuite struct {
	suite.Suite
	pgt        *dbtest.PostgresTest
	ctx        context.Context
	repository domain.Repository
}

func TestAuth(t *testing.T) {
	suite.Run(t, &AuthTestSuite{})
}

func (a *AuthTestSuite) SetupSuite() {
	a.ctx = context.Background()

	var err error
	a.pgt, err = dbtest.NewPostgresTest(a.ctx, zap.L(), "../../../database/migrations", &dbpool.Config{
		Logging:      false,
		Host:         "localhost",
		Port:         "5432",
		User:         "testuser",
		Password:     "1234",
		DatabaseName: "authtestdb",
		MaxConns:     1,
		MinConns:     1,
	})
	a.Require().NoError(err)

	err = a.pgt.MigrateUp()
	a.Require().NoError(err)

	a.repository = New(a.pgt.DB())
}

func (a *AuthTestSuite) TearDownSuite() {
	a.Require().NoError(a.pgt.TearDown())
}

func (a *AuthTestSuite) TearDownTest() {
	_, err := a.pgt.DB().Exec(a.ctx, "TRUNCATE TABLE auth_identities, users CASCADE;")
	a.Require().NoError(err)
}

func (a *AuthTestSuite) TestCreateAuthIdentity() {
	ctx := a.ctx
	t := a.T()

	// Generate and insert two mock users
	mockUsers := usertestutils.GenerateMockUsers(2)
	testutils.InsertMockUsersIntoDB(t, a.pgt.DB(), ctx, mockUsers)

	// Generate auth identity params for both users
	mockAuthParams := testutils.GenerateMockAuthParams(mockUsers)

	t.Run("Create Valid Auth Identity", func(t *testing.T) {
		authParams := mockAuthParams[0]

		authIdentity, err := a.repository.CreateAuthIdentity(ctx, authParams)
		a.Require().NoError(err)
		a.Equal(authParams.AuthID, authIdentity.AuthID)
		a.Equal(authParams.Provider, authIdentity.Provider)
		a.Equal(authParams.UserID, authIdentity.UserID)
		a.Equal(authParams.Role, authIdentity.Role)
		a.NotNil(authIdentity.CreatedAt)
		a.NotNil(authIdentity.UpdatedAt)
	})

	t.Run("Duplicate Auth ID", func(t *testing.T) {
		authParams := mockAuthParams[1]

		// First insert
		_, err := a.repository.CreateAuthIdentity(ctx, authParams)
		a.Require().NoError(err)

		// Attempt duplicate insert
		_, err = a.repository.CreateAuthIdentity(ctx, authParams)
		a.Require().Error(err)
		a.ErrorIs(err, ErrAuthIDAlreadyExists)
	})
}

func (a *AuthTestSuite) TestGetAuthIdentityByAuthID() {
	ctx := a.ctx
	t := a.T()

	// Generate and insert one mock user
	mockUsers := usertestutils.GenerateMockUsers(1)
	testutils.InsertMockUsersIntoDB(t, a.pgt.DB(), ctx, mockUsers)

	// Generate auth identity params for the mock user
	mockAuthParams := testutils.GenerateMockAuthParams(mockUsers)
	authParams := mockAuthParams[0]

	t.Run("Retrieve existing auth identity by AuthID", func(t *testing.T) {
		// Create auth identity
		created, err := a.repository.CreateAuthIdentity(ctx, authParams)
		a.Require().NoError(err)

		// Act
		result, err := a.repository.GetAuthIdentityByAuthID(ctx, authParams.AuthID)

		// Assert
		a.Require().NoError(err)
		a.Equal(created.AuthID, result.AuthID)
		a.Equal(created.Provider, result.Provider)
		a.Equal(created.UserID, result.UserID)
		a.Equal(created.Role, result.Role)
	})

	t.Run("Return error if auth ID does not exist", func(t *testing.T) {
		invalidAuthID := "auth0|does-not-exist"

		_, err := a.repository.GetAuthIdentityByAuthID(ctx, invalidAuthID)

		a.Require().Error(err)
		a.ErrorIs(err, common.ErrNotFound)
	})
}

// TestGetAuthIdentityByUserID tests the GetAuthIdentityByUserID method.
func (a *AuthTestSuite) TestGetAuthIdentityByUserID() {
	ctx := a.ctx
	t := a.T()

	// Generate and insert one mock user
	mockUsers := usertestutils.GenerateMockUsers(1)
	testutils.InsertMockUsersIntoDB(t, a.pgt.DB(), ctx, mockUsers)

	// Generate auth identity params for the mock user
	mockAuthParams := testutils.GenerateMockAuthParams(mockUsers)
	authParams := mockAuthParams[0]

	t.Run("Retrieve existing auth identity by UserID", func(t *testing.T) {
		// Create auth identity
		created, err := a.repository.CreateAuthIdentity(ctx, authParams)
		a.Require().NoError(err)

		// Act
		result, err := a.repository.GetAuthIdentityByUserID(ctx, authParams.UserID)

		// Assert
		a.Require().NoError(err)
		a.Equal(created.AuthID, result.AuthID)
		a.Equal(created.Provider, result.Provider)
		a.Equal(created.UserID, result.UserID)
		a.Equal(created.Role, result.Role)
	})

	t.Run("Return error if user ID does not exist", func(t *testing.T) {
		randomUserID := uuid.New()

		_, err := a.repository.GetAuthIdentityByUserID(ctx, randomUserID)

		a.Require().Error(err)
		a.ErrorIs(err, common.ErrNotFound)
	})
}

// TestUpdateAuthIdentityRole tests the UpdateAuthIdentityRole method.
func (a *AuthTestSuite) TestUpdateAuthIdentityRole() {
	ctx := a.ctx
	t := a.T()

	// Generate and insert one mock user
	mockUsers := usertestutils.GenerateMockUsers(1)
	testutils.InsertMockUsersIntoDB(t, a.pgt.DB(), ctx, mockUsers)

	// Generate auth identity params for the mock user
	mockAuthParams := testutils.GenerateMockAuthParams(mockUsers)
	authParams := mockAuthParams[0]

	t.Run("Update role for existing auth identity", func(t *testing.T) {
		// Create auth identity
		created, err := a.repository.CreateAuthIdentity(ctx, authParams)
		a.Require().NoError(err)

		// Prepare update params
		updateParams := domain.UpdateAuthIdentityParams{
			AuthID: created.AuthID,
			Role:   "admin", // New role
		}

		// Act
		updated, err := a.repository.UpdateAuthIdentityRole(ctx, updateParams)

		// Assert
		a.Require().NoError(err)
		a.Equal(created.AuthID, updated.AuthID)
		a.Equal(created.Provider, updated.Provider)
		a.Equal(created.UserID, updated.UserID)
		a.Equal(updateParams.Role, updated.Role) // Role should be updated
	})

	t.Run("Return error if auth ID does not exist", func(t *testing.T) {
		updateParams := domain.UpdateAuthIdentityParams{
			AuthID: "auth0|nonexistent123",
			Role:   "admin",
		}

		_, err := a.repository.UpdateAuthIdentityRole(ctx, updateParams)

		a.Require().Error(err)
		a.ErrorIs(err, common.ErrNotFound)
	})
}

// TestDeleteAuthIdentityByAuthID tests the DeleteAuthIdentityByAuthID method.
func (a *AuthTestSuite) TestDeleteAuthIdentityByAuthID() {
	ctx := a.ctx
	t := a.T()

	// Generate and insert one mock user
	mockUsers := usertestutils.GenerateMockUsers(1)
	testutils.InsertMockUsersIntoDB(t, a.pgt.DB(), ctx, mockUsers)

	// Generate auth identity params for the mock user
	mockAuthParams := testutils.GenerateMockAuthParams(mockUsers)
	authParams := mockAuthParams[0]

	t.Run("Delete existing auth identity by AuthID", func(t *testing.T) {
		// Create auth identity
		created, err := a.repository.CreateAuthIdentity(ctx, authParams)
		a.Require().NoError(err)

		// Act: Delete the auth identity
		err = a.repository.DeleteAuthIdentityByAuthID(ctx, created.AuthID)

		// Assert
		a.Require().NoError(err)

		// Verify deletion: Try to fetch it and expect not found
		_, err = a.repository.GetAuthIdentityByAuthID(ctx, created.AuthID)
		a.Require().Error(err)
		a.ErrorIs(err, common.ErrNotFound)
	})

	t.Run("Return error if auth ID does not exist", func(t *testing.T) {
		invalidAuthID := "auth0|nonexistent123"

		// Act
		err := a.repository.DeleteAuthIdentityByAuthID(ctx, invalidAuthID)

		// Assert
		a.Require().Error(err)
		a.ErrorIs(err, common.ErrNotFound)
	})
}
