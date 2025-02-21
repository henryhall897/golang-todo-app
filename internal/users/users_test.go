//go:build unit

package users

import (
	"context"
	"fmt"
	"testing"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/core/dbpool"
	"github.com/henryhall897/golang-todo-app/pkg/dbtest"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type UserTestSuite struct {
	suite.Suite
	pgt   *dbtest.PostgresTest
	ctx   context.Context
	store *UserStore
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

func (u *UserTestSuite) CreateSampleUsers(ctx context.Context, count int) ([]User, error) {
	var users []User
	for i := 1; i <= count; i++ {
		name := fmt.Sprintf("Joe %d", i)
		email := fmt.Sprintf("joe%d@example.com", i)
		newUser := CreateUserParams{
			Name:  name,
			Email: email,
		}
		user, err := u.store.CreateUser(ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create sample user %d: %w", i, err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (u *UserTestSuite) TestCreateUser() {
	ctx := u.ctx
	t := u.T() // Get the testing instance

	t.Run("Create Valid User", func(t *testing.T) {
		// Arrange - Prepare new user data
		newUser := CreateUserParams{
			Name:  "John Doe",
			Email: "john.doe@example.com",
		}

		// Act - Create user
		createdUser, err := u.store.CreateUser(ctx, newUser)

		// Assert - Validate creation
		u.Require().NoError(err)
		u.Require().NotNil(createdUser)
		u.Equal(newUser.Name, createdUser.Name)
		u.Equal(newUser.Email, createdUser.Email)

		// Verify user exists in the database
		retrievedUser, err := u.store.GetUserByID(ctx, createdUser.ID)
		u.Require().NoError(err)
		u.Require().NotNil(retrievedUser)
		u.Equal(createdUser.ID, retrievedUser.ID)
		u.Equal(createdUser.Name, retrievedUser.Name)
		u.Equal(createdUser.Email, retrievedUser.Email)
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		// Arrange - Create first user
		duplicateUser := CreateUserParams{
			Name:  "John Smith",
			Email: "duplicate@example.com",
		}

		_, err := u.store.CreateUser(ctx, duplicateUser)
		u.Require().NoError(err)

		// Act - Try creating user with same email
		_, err = u.store.CreateUser(ctx, duplicateUser)

		// Assert - Expect duplicate key error
		u.Require().Error(err)
		u.Contains(err.Error(), "duplicate key value violates unique constraint")
	})
}

// TestGetUserByID validates retrieving a user by ID
func (u *UserTestSuite) TestGetUserByID() {
	ctx := u.ctx

	t := u.T() // Get the underlying testing.T instance

	t.Run("Valid User ID", func(t *testing.T) {
		// Arrange - Create a sample user
		users, err := u.CreateSampleUsers(ctx, 1)
		u.Require().NoError(err)
		createdUser := users[0]

		// Act
		retrievedUser, err := u.store.GetUserByID(ctx, createdUser.ID)

		// Assert
		u.Require().NoError(err)
		u.Require().NotNil(retrievedUser)
		u.Equal(createdUser.ID, retrievedUser.ID)
		u.Equal(createdUser.Name, retrievedUser.Name)
		u.Equal(createdUser.Email, retrievedUser.Email)
	})

	t.Run("User Not Found", func(t *testing.T) {
		// Act
		nonExistentID := uuid.New()
		_, err := u.store.GetUserByID(ctx, nonExistentID)

		// Assert
		u.Require().Error(err)
		u.ErrorIs(err, common.ErrNotFound)
	})
}

// TestGetUserByEmail validates retrieving a user by email
func (u *UserTestSuite) TestGetUserByEmail() {
	ctx := u.ctx
	t := u.T() // Get the testing instance

	t.Run("Valid User Email", func(t *testing.T) {
		// Arrange - Create a sample user
		users, err := u.CreateSampleUsers(ctx, 1)
		u.Require().NoError(err)
		createdUser := users[0]

		// Act
		retrievedUser, err := u.store.GetUserByEmail(ctx, createdUser.Email)

		// Assert
		u.Require().NoError(err)
		u.Require().NotNil(retrievedUser)
		u.Equal(createdUser.ID, retrievedUser.ID)
		u.Equal(createdUser.Name, retrievedUser.Name)
		u.Equal(createdUser.Email, retrievedUser.Email)
	})

	t.Run("User Not Found", func(t *testing.T) {
		// Arrange
		nonExistentEmail := "nonexistent@example.com"

		// Act
		_, err := u.store.GetUserByEmail(ctx, nonExistentEmail)

		// Assert
		u.Require().Error(err)
		u.ErrorIs(err, common.ErrNotFound)
	})
}

// TestListUsers checks listing all users
func (u *UserTestSuite) TestListUsers() {
	ctx := u.ctx
	t := u.T() // Get the testing instance
	users, err := u.CreateSampleUsers(ctx, 3)

	t.Run("List All Users", func(t *testing.T) {
		// Arrange - Create 3 sample users

		u.Require().NoError(err)
		u.Require().Len(users, 3, "Expected 3 users to be created")

		// Act - List users with pagination
		params := ListUsersParams{
			Limit:  3,
			Offset: 0,
		}
		results, err := u.store.ListUsers(ctx, params)

		// Assert
		u.Require().NoError(err)
		u.Require().Len(results, len(users), "Expected number of users to match")

		// Reverse the expected order to match descending order by created_at
		expectedOrder := []User{users[2], users[1], users[0]} // Latest user first

		for i, result := range results {
			u.Equal(expectedOrder[i].Name, result.Name, "Mismatched user name at index %d", i)
			u.Equal(expectedOrder[i].Email, result.Email, "Mismatched user email at index %d", i)
		}
	})

	t.Run("List Users With Pagination", func(t *testing.T) {
		u.Require().NoError(err)
		u.Require().Len(users, 3, "Expected 3 users to be created")

		// Act - Retrieve only 2 users, skipping the latest user
		params := ListUsersParams{
			Limit:  2,
			Offset: 1, // Skip the most recently created user
		}
		results, err := u.store.ListUsers(ctx, params)

		// Assert
		u.Require().NoError(err)
		u.Require().Len(results, 2)

		// Expected order: users[1], users[0] (Skipping users[2] due to offset)
		expectedOrder := []User{users[1], users[0]}

		for i, result := range results {
			u.Equal(expectedOrder[i].Name, result.Name, "Mismatched user name at index %d", i)
			u.Equal(expectedOrder[i].Email, result.Email, "Mismatched user email at index %d", i)
		}
	})

	t.Run("List Users With Empty Results", func(t *testing.T) {
		// Arrange - Create 3 sample users
		u.Require().NoError(err)

		// Act with OFFSET exceeding total rows
		params := ListUsersParams{
			Limit:  2,
			Offset: 10, // Skip more rows than exist
		}
		results, err := u.store.ListUsers(ctx, params)

		// Assert
		u.Require().NoError(err)
		u.Require().Empty(results, "Expected no results when offset exceeds row count")

		// Act with LIMIT = 0
		params = ListUsersParams{
			Limit:  0, // No rows should be returned
			Offset: 0,
		}
		results, err = u.store.ListUsers(ctx, params)

		// Assert
		u.Require().NoError(err)
		u.Require().Empty(results, "Expected no results when limit is 0")
	})
}

func (u *UserTestSuite) TestUpdateUser() {
	ctx := u.ctx
	t := u.T() // Get the testing instance

	// Arrange - Create a sample user
	users, err := u.CreateSampleUsers(ctx, 1)
	u.Require().NoError(err)
	createdUser := users[0]

	// Update user details
	updatedName := "Jane Smith"
	updatedEmail := "jane.smith@example.com"

	t.Run("Update User Name and Email", func(t *testing.T) {
		updateParams := UpdateUserParams{
			ID:    createdUser.ID,
			Name:  updatedName,
			Email: updatedEmail,
		}

		// Act - Update the user
		updatedUser, err := u.store.UpdateUser(ctx, updateParams)

		// Assert - Verify update
		u.Require().NoError(err, "Failed to update user")
		u.Require().NotNil(updatedUser)
		u.Equal(createdUser.ID, updatedUser.ID, "User ID should remain unchanged")
		u.Equal(updatedName, updatedUser.Name, "User name should be updated")
		u.Equal(updatedEmail, updatedUser.Email, "User email should be updated")

		// Act - Retrieve the updated user to confirm changes in the database
		retrievedUser, err := u.store.GetUserByID(ctx, createdUser.ID)

		// Assert - Verify retrieved user matches updated details
		u.Require().NoError(err, "Failed to retrieve updated user")
		u.Require().NotNil(retrievedUser)
		u.Equal(updatedName, retrievedUser.Name, "Retrieved user name should match the updated name")
		u.Equal(updatedEmail, retrievedUser.Email, "Retrieved user email should match the updated email")
	})

	t.Run("Update User with Partial Fields", func(t *testing.T) {
		// Arrange - Only update the name
		partialUpdatedName := "Jane Partial"
		partialUpdateParams := UpdateUserParams{
			ID:    createdUser.ID,
			Name:  partialUpdatedName,
			Email: updatedEmail, // Email remains unchanged
		}

		// Act - Partial update
		updatedUser, err := u.store.UpdateUser(ctx, partialUpdateParams)

		// Assert - Ensure only the name is updated
		u.Require().NoError(err)
		u.Require().NotNil(updatedUser)
		u.Equal(partialUpdatedName, updatedUser.Name, "User name should be updated")
		u.Equal(updatedEmail, updatedUser.Email, "User email should remain unchanged")

		// Act - Retrieve and verify changes in the database
		retrievedUser, err := u.store.GetUserByID(ctx, createdUser.ID)

		// Assert - Verify retrieved user matches updated details
		u.Require().NoError(err)
		u.Require().NotNil(retrievedUser)
		u.Equal(partialUpdatedName, retrievedUser.Name, "Retrieved user name should match the updated name")
		u.Equal(updatedEmail, retrievedUser.Email, "Retrieved user email should remain unchanged")
	})
}

func (u *UserTestSuite) TestDeleteUser() {
	ctx := u.ctx
	t := u.T() // Get the testing instance

	t.Run("Delete Existing User", func(t *testing.T) {
		// Arrange - Create a sample user
		users, err := u.CreateSampleUsers(ctx, 1)
		u.Require().NoError(err)
		createdUser := users[0]

		// Act - Delete the user
		err = u.store.DeleteUser(ctx, createdUser.ID)

		// Assert - Verify deletion
		u.Require().NoError(err, "Failed to delete user")

		// Act - Try retrieving the deleted user
		_, err = u.store.GetUserByID(ctx, createdUser.ID)

		// Assert - User should no longer exist
		u.Require().Error(err)
		u.ErrorIs(err, common.ErrNotFound, "not found")
	})

	t.Run("Delete Non-Existent User", func(t *testing.T) {
		// Arrange - Generate a random UUID
		nonExistentID := uuid.New()

		// Act - Try deleting a user that doesn't exist
		err := u.store.DeleteUser(ctx, nonExistentID)

		// Assert - Should return ErrNotFound
		u.Require().Error(err)
		u.ErrorIs(err, common.ErrNotFound, "Expected ErrNotFound for non-existent user")
	})
}
