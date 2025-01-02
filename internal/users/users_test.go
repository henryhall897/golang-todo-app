//go:build unit

package users

import (
	"context"
	"testing"

	"golang-todo-app/internal/core/common"
	"golang-todo-app/internal/core/dbpool"
	"golang-todo-app/internal/users/gen"
	"golang-todo-app/pkg/dbtest"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type UserTestSuite struct {
	suite.Suite
	pgt   *dbtest.PostgresTest
	ctx   context.Context
	store *Store
}

func TestUsers(t *testing.T) {
	suite.Run(t, &UserTestSuite{})
}

func (u *UserTestSuite) SetupSuite() {
	u.ctx = context.Background()

	var err error
	u.pgt, err = dbtest.NewPostgresTest(u.ctx, zap.L(), "../../migrations", &dbpool.Config{
		Logging:      false,
		Host:         "localhost",
		Port:         "5432",
		User:         "testuser",
		Password:     "1234",
		DatabaseName: "usertestdb",
		MaxConns:     1,
		MinConns:     1,
	})
	u.Require().NoError(err)

	err = u.pgt.MigrateUp()
	u.Require().NoError(err)

	u.store = New(u.pgt.DB())
}

func (u *UserTestSuite) TearDownSuite() {
	u.Require().NoError(u.pgt.TearDown())
}

func (u *UserTestSuite) TearDownTest() {
	_, err := u.pgt.DB().Exec(u.ctx, "TRUNCATE TABLE users CASCADE;")
	u.Require().NoError(err)
}

// TestCreateUser validates the creation of a new user
func (u *UserTestSuite) TestCreateUser() {
	ctx := u.ctx

	// Arrange
	name := "John Doe"
	email := "john.doe@example.com"

	// Act
	createdUser, err := u.store.CreateUser(ctx, name, email)

	// Assert
	u.Require().NoError(err)
	u.Require().NotNil(createdUser)
	u.Equal(name, createdUser.Name)
	u.Equal(email, createdUser.Email)

	// Verify user exists in the database
	retrievedUser, err := u.store.GetUserByID(ctx, createdUser.ID)
	u.Require().NoError(err)
	u.Require().NotNil(retrievedUser)
	u.Equal(createdUser.ID, retrievedUser.ID)
	u.Equal(name, retrievedUser.Name)
	u.Equal(email, retrievedUser.Email)
}

// TestGetUserByID validates retrieving a user by ID
func (u *UserTestSuite) TestGetUserByID() {
	ctx := u.ctx

	// Arrange
	name := "Jane Doe"
	email := "jane.doe@example.com"
	createdUser, err := u.store.CreateUser(ctx, name, email)
	u.Require().NoError(err)

	// Act
	retrievedUser, err := u.store.GetUserByID(ctx, createdUser.ID)

	// Assert
	u.Require().NoError(err)
	u.Require().NotNil(retrievedUser)
	u.Equal(createdUser.ID, retrievedUser.ID)
	u.Equal(name, retrievedUser.Name)
	u.Equal(email, retrievedUser.Email)
}

// TestGetUserByIDNotFound checks behavior when user is not found
func (u *UserTestSuite) TestGetUserByIDNotFound() {
	ctx := u.ctx

	// Act
	nonExistentID := uuid.New()
	_, err := u.store.GetUserByID(ctx, nonExistentID)

	// Assert
	u.Require().Error(err)
	u.ErrorIs(err, common.ErrNotFound)
}

// TestCreateUserDuplicateEmail checks duplicate email constraint
func (u *UserTestSuite) TestCreateUserDuplicateEmail() {
	ctx := u.ctx

	// Arrange
	name1 := "John Smith"
	name2 := "Jane Smith"
	email := "duplicate@example.com"

	_, err := u.store.CreateUser(ctx, name1, email)
	u.Require().NoError(err)

	// Act
	_, err = u.store.CreateUser(ctx, name2, email)

	// Assert
	u.Require().Error(err)
	u.Contains(err.Error(), "duplicate key value violates unique constraint")
}

// TestGetUserByEmail validates retrieving a user by email
func (u *UserTestSuite) TestGetUserByEmail() {
	ctx := u.ctx

	// Arrange
	name := "John Doe"
	email := "john.doe@example.com"

	// Create a user in the database
	createdUser, err := u.store.CreateUser(ctx, name, email)
	u.Require().NoError(err)

	// Act
	retrievedUser, err := u.store.GetUserByEmail(ctx, email)

	// Assert
	u.Require().NoError(err)
	u.Require().NotNil(retrievedUser)
	u.Equal(createdUser.ID, retrievedUser.ID)
	u.Equal(name, retrievedUser.Name)
	u.Equal(email, retrievedUser.Email)
}

// TestGetUserByEmailNotFound checks behavior when email is not found
func (u *UserTestSuite) TestGetUserByEmailNotFound() {
	ctx := u.ctx

	// Arrange
	nonExistentEmail := "nonexistent@example.com"

	// Act
	_, err := u.store.GetUserByEmail(ctx, nonExistentEmail)

	// Assert
	u.Require().Error(err)
	u.ErrorIs(err, common.ErrNotFound)
}

// TestListUsers checks listing all users
func (u *UserTestSuite) TestListUsers() {
	ctx := u.ctx

	// Arrange
	users := []struct {
		Name  string
		Email string
	}{
		{"Alice", "alice@example.com"},
		{"Bob", "bob@example.com"},
		{"Charlie", "charlie@example.com"},
	}

	// Create users in the order they are defined
	for _, user := range users {
		_, err := u.store.CreateUser(ctx, user.Name, user.Email)
		u.Require().NoError(err, "Failed to create user: %s", user.Name)
	}

	// Act
	params := gen.ListUsersParams{
		Limit:  3,
		Offset: 0,
	}
	results, err := u.store.ListUsers(ctx, params)

	// Assert
	u.Require().NoError(err)
	u.Require().Len(results, len(users))

	// Reverse the expected order to match descending order by created_at
	expectedOrder := []struct {
		Name  string
		Email string
	}{
		{"Charlie", "charlie@example.com"},
		{"Bob", "bob@example.com"},
		{"Alice", "alice@example.com"},
	}

	for i, result := range results {
		u.Equal(expectedOrder[i].Name, result.Name, "Mismatched user name")
		u.Equal(expectedOrder[i].Email, result.Email, "Mismatched user email")
	}
}

// Tests if list user works with pagination.
func (u *UserTestSuite) TestListUsersWithPagination() {
	ctx := u.ctx

	// Arrange
	users := []struct {
		Name  string
		Email string
	}{
		{"Alice", "alice@example.com"},
		{"Bob", "bob@example.com"},
		{"Charlie", "charlie@example.com"},
	}

	for _, user := range users {
		_, err := u.store.CreateUser(ctx, user.Name, user.Email)
		u.Require().NoError(err, "Failed to create user: %s", user.Name)
	}

	// Act
	params := gen.ListUsersParams{
		Limit:  2, // Retrieve only 2 users
		Offset: 1, // Skip the first user (Charlie)
	}
	results, err := u.store.ListUsers(ctx, params)

	// Assert
	u.Require().NoError(err)
	u.Require().Len(results, 2)

	// Expected order: Bob, Alice (skipping Charlie due to offset)
	expectedOrder := []struct {
		Name  string
		Email string
	}{
		{"Bob", "bob@example.com"},
		{"Alice", "alice@example.com"},
	}

	for i, result := range results {
		u.Equal(expectedOrder[i].Name, result.Name, "Mismatched user name")
		u.Equal(expectedOrder[i].Email, result.Email, "Mismatched user email")
	}
}

// Tests Empty Results.
func (u *UserTestSuite) TestListUsersEmptyResults() {
	ctx := u.ctx

	// Arrange
	users := []struct {
		Name  string
		Email string
	}{
		{"Alice", "alice@example.com"},
		{"Bob", "bob@example.com"},
		{"Charlie", "charlie@example.com"},
	}

	for _, user := range users {
		_, err := u.store.CreateUser(ctx, user.Name, user.Email)
		u.Require().NoError(err, "Failed to create user: %s", user.Name)
	}

	// Act with OFFSET exceeding total rows
	params := gen.ListUsersParams{
		Limit:  2,
		Offset: 10, // Skip more rows than exist
	}
	results, err := u.store.ListUsers(ctx, params)

	// Assert
	u.Require().NoError(err)
	u.Require().Empty(results, "Expected no results when offset exceeds row count")

	// Act with LIMIT = 0
	params = gen.ListUsersParams{
		Limit:  0, // No rows should be returned
		Offset: 0,
	}
	results, err = u.store.ListUsers(ctx, params)

	// Assert
	u.Require().NoError(err)
	u.Require().Empty(results, "Expected no results when limit is 0")
}

func (u *UserTestSuite) TestUpdateUser() {
	ctx := u.ctx

	// Arrange
	name := "Jane Doe"
	email := "jane.doe@example.com"
	createdUser, err := u.store.CreateUser(ctx, name, email)
	u.Require().NoError(err, "Failed to create user")

	updatedName := "Jane Smith"
	updatedEmail := "jane.smith@example.com"

	// Act: Update the user's name and email
	updatedUser, err := u.store.UpdateUser(ctx, createdUser.ID, updatedName, updatedEmail)

	// Assert: Verify the updated user details
	u.Require().NoError(err, "Failed to update user")
	u.Require().NotNil(updatedUser)
	u.Equal(createdUser.ID, updatedUser.ID, "User ID should remain unchanged")
	u.Equal(updatedName, updatedUser.Name, "User name should be updated")
	u.Equal(updatedEmail, updatedUser.Email, "User email should be updated")

	// Act: Retrieve the updated user to confirm changes in the database
	retrievedUser, err := u.store.GetUserByID(ctx, createdUser.ID)

	// Assert: Verify the retrieved user matches the updated details
	u.Require().NoError(err, "Failed to retrieve updated user")
	u.Require().NotNil(retrievedUser)
	u.Equal(updatedName, retrievedUser.Name, "Retrieved user name should match the updated name")
	u.Equal(updatedEmail, retrievedUser.Email, "Retrieved user email should match the updated email")
}

func (u *UserTestSuite) TestDeleteUser() {
	ctx := u.ctx

	// Arrange: Create a user to delete
	name := "John Doe"
	email := "john.doe@example.com"
	createdUser, err := u.store.CreateUser(ctx, name, email)
	u.Require().NoError(err, "Failed to create user")

	// Act: Delete the user
	err = u.store.DeleteUser(ctx, createdUser.ID)

	// Assert: Verify the user was deleted successfully
	u.Require().NoError(err, "Failed to delete user")

	// Act: Try retrieving the deleted user
	_, err = u.store.GetUserByID(ctx, createdUser.ID)

	// Assert: Verify the user no longer exists
	u.Require().Error(err)
	u.ErrorIs(err, common.ErrNotFound, "Expected ErrNotFound for deleted user")
}

func (u *UserTestSuite) TestDeleteNonExistentUser() {
	ctx := u.ctx

	// Arrange: Generate a random UUID
	nonExistentID := uuid.New()

	// Act: Try deleting a user that doesn't exist
	err := u.store.DeleteUser(ctx, nonExistentID)

	// Assert: Verify it returns ErrNotFound
	u.Require().Error(err)
	u.ErrorIs(err, common.ErrNotFound, "Expected ErrNotFound for non-existent user")
}
