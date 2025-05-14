package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/core/dbpool"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
	"github.com/henryhall897/golang-todo-app/internal/users/testutils"
	"github.com/henryhall897/golang-todo-app/pkg/dbtest"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type UserTestSuite struct {
	suite.Suite
	pgt        *dbtest.PostgresTest
	ctx        context.Context
	repository domain.Repository
}

func TestUsers(t *testing.T) {
	suite.Run(t, &UserTestSuite{})
}

func (u *UserTestSuite) SetupSuite() {
	u.ctx = context.Background()

	var err error
	u.pgt, err = dbtest.NewPostgresTest(u.ctx, zap.L(), "../../../database/migrations", &dbpool.Config{
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

	u.repository = New(u.pgt.DB())
}

func (u *UserTestSuite) TearDownSuite() {
	u.Require().NoError(u.pgt.TearDown())
}

func (u *UserTestSuite) TearDownTest() {
	_, err := u.pgt.DB().Exec(u.ctx, "TRUNCATE TABLE users CASCADE;")
	u.Require().NoError(err)
}

func (u *UserTestSuite) CreateSampleUsers(ctx context.Context, count int) ([]domain.User, error) {
	// Generate mock users with the utility function
	mockUsers := testutils.GenerateMockUsers(count)
	var createdUsers []domain.User

	// Insert each mock user into the database
	for i, mockUser := range mockUsers {
		// Convert mock user to creation parameters
		newUser := domain.CreateUserParams{
			Name:  mockUser.Name,
			Email: mockUser.Email,
			Role:  mockUser.Role,
		}

		// Create the user in the database
		user, err := u.repository.CreateUser(ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create sample user %d: %w", i+1, err)
		}

		createdUsers = append(createdUsers, user)
	}

	return createdUsers, nil
}

func (u *UserTestSuite) TestCreateUser() {
	ctx := u.ctx
	t := u.T() // Get the testing instance

	t.Run("Create Valid domain.User", func(t *testing.T) {
		newUser := domain.CreateUserParams{
			Name:  "John Doe",
			Email: "john.doe@example.com",
			Role:  "user",
		}

		createdUser, err := u.repository.CreateUser(ctx, newUser)

		u.Require().NoError(err)
		u.Require().NotNil(createdUser)
		u.Equal(newUser.Name, createdUser.Name)
		u.Equal(newUser.Email, createdUser.Email)

		retrievedUser, err := u.repository.GetUserByID(ctx, createdUser.ID)
		u.Require().NoError(err)
		u.Require().Equal(createdUser.ID, retrievedUser.ID)
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		// Arrange
		original := domain.CreateUserParams{
			Name:  "Jane Smith",
			Email: "jane@example.com",
			Role:  "admin",
		}
		duplicate := domain.CreateUserParams{
			Name:  "Fake Jane",
			Email: "jane@example.com", // same email
			Role:  "user",
		}

		_, err := u.repository.CreateUser(ctx, original)
		u.Require().NoError(err)

		_, err = u.repository.CreateUser(ctx, duplicate)
		u.Require().Error(err)
		u.ErrorIs(err, ErrEmailAlreadyExists)
	})
}

// TestGetUserByID validates retrieving a user by ID
func (u *UserTestSuite) TestGetUserByID() {
	ctx := u.ctx

	t := u.T() // Get the underlying testing.T instance

	t.Run("Valid domain.User ID", func(t *testing.T) {
		// Arrange - Create a sample user
		users, err := u.CreateSampleUsers(ctx, 1)
		u.Require().NoError(err)
		createdUser := users[0]

		// Act
		retrievedUser, err := u.repository.GetUserByID(ctx, createdUser.ID)

		// Assert
		u.Require().NoError(err)
		u.Require().NotNil(retrievedUser)
		u.Equal(createdUser.ID, retrievedUser.ID)
		u.Equal(createdUser.Name, retrievedUser.Name)
		u.Equal(createdUser.Email, retrievedUser.Email)
	})

	t.Run("domain.User Not Found", func(t *testing.T) {
		// Act
		nonExistentID := uuid.New()
		_, err := u.repository.GetUserByID(ctx, nonExistentID)

		// Assert
		u.Require().Error(err)
		u.ErrorIs(err, common.ErrNotFound)
	})
}

// TestGetUserByEmail validates retrieving a user by email
func (u *UserTestSuite) TestGetUserByEmail() {
	ctx := u.ctx
	t := u.T() // Get the testing instance

	t.Run("Valid domain.User Email", func(t *testing.T) {
		// Arrange - Create a sample user
		users, err := u.CreateSampleUsers(ctx, 1)
		u.Require().NoError(err)
		createdUser := users[0]

		// Act
		retrievedUser, err := u.repository.GetUserByEmail(ctx, createdUser.Email)

		// Assert
		u.Require().NoError(err)
		u.Require().NotNil(retrievedUser)
		u.Equal(createdUser.ID, retrievedUser.ID)
		u.Equal(createdUser.Name, retrievedUser.Name)
		u.Equal(createdUser.Email, retrievedUser.Email)
	})

	t.Run("domain.User Not Found", func(t *testing.T) {
		// Arrange
		nonExistentEmail := "nonexistent@example.com"

		// Act
		_, err := u.repository.GetUserByEmail(ctx, nonExistentEmail)

		// Assert
		u.Require().Error(err)
		u.ErrorIs(err, common.ErrNotFound)
	})
}

// TestGetUsers checks listing all users
func (u *UserTestSuite) TestGetUsers() {
	ctx := u.ctx
	t := u.T() // Get the testing instance
	users, err := u.CreateSampleUsers(ctx, 3)

	t.Run("List All Users", func(t *testing.T) {
		// Arrange - Create 3 sample users

		u.Require().NoError(err)
		u.Require().Len(users, 3, "Expected 3 users to be created")

		// Act - List users with pagination
		params := domain.GetUsersParams{
			Limit:  3,
			Offset: 0,
		}
		results, err := u.repository.GetUsers(ctx, params)

		// Assert
		u.Require().NoError(err)
		u.Require().Len(results, len(users), "Expected number of users to match")

		// Reverse the expected order to match descending order by created_at
		expectedOrder := []domain.User{users[2], users[1], users[0]} // Latest user first

		for i, result := range results {
			u.Equal(expectedOrder[i].Name, result.Name, "Mismatched user name at index %d", i)
			u.Equal(expectedOrder[i].Email, result.Email, "Mismatched user email at index %d", i)
		}
	})

	t.Run("List Users With Pagination", func(t *testing.T) {
		u.Require().NoError(err)
		u.Require().Len(users, 3, "Expected 3 users to be created")

		// Act - Retrieve only 2 users, skipping the latest user
		params := domain.GetUsersParams{
			Limit:  2,
			Offset: 1, // Skip the most recently created user
		}
		results, err := u.repository.GetUsers(ctx, params)

		// Assert
		u.Require().NoError(err)
		u.Require().Len(results, 2)

		// Expected order: users[1], users[0] (Skipping users[2] due to offset)
		expectedOrder := []domain.User{users[1], users[0]}

		for i, result := range results {
			u.Equal(expectedOrder[i].Name, result.Name, "Mismatched user name at index %d", i)
			u.Equal(expectedOrder[i].Email, result.Email, "Mismatched user email at index %d", i)
		}
	})

	t.Run("List Users With Empty Results", func(t *testing.T) {
		// Arrange - Create 3 sample users
		u.Require().NoError(err)

		// Act with OFFSET exceeding total rows
		params := domain.GetUsersParams{
			Limit:  2,
			Offset: 10, // Skip more rows than exist
		}
		results, err := u.repository.GetUsers(ctx, params)

		// Assert
		u.Require().NoError(err)
		u.Require().Empty(results, "Expected no results when offset exceeds row count")

		// Act with LIMIT = 0
		params = domain.GetUsersParams{
			Limit:  0, // No rows should be returned
			Offset: 0,
		}
		results, err = u.repository.GetUsers(ctx, params)

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

	t.Run("Update domain.User Name and Email", func(t *testing.T) {
		updateParams := domain.UpdateUserParams{
			ID:    createdUser.ID,
			Name:  updatedName,
			Email: updatedEmail,
		}

		// Act - Update the user
		updatedUser, err := u.repository.UpdateUser(ctx, updateParams)

		// Assert - Verify update
		u.Require().NoError(err, "Failed to update user")
		u.Require().NotNil(updatedUser)
		u.Equal(createdUser.ID, updatedUser.ID, "domain.User ID should remain unchanged")
		u.Equal(updatedName, updatedUser.Name, "domain.User name should be updated")
		u.Equal(updatedEmail, updatedUser.Email, "domain.User email should be updated")

		// Act - Retrieve the updated user to confirm changes in the database
		retrievedUser, err := u.repository.GetUserByID(ctx, createdUser.ID)

		// Assert - Verify retrieved user matches updated details
		u.Require().NoError(err, "Failed to retrieve updated user")
		u.Require().NotNil(retrievedUser)
		u.Equal(updatedName, retrievedUser.Name, "Retrieved user name should match the updated name")
		u.Equal(updatedEmail, retrievedUser.Email, "Retrieved user email should match the updated email")
	})
	t.Run("Update domain.User with Partial Fields", func(t *testing.T) {
		// Arrange - Only update the name
		partialUpdatedName := "Jane Partial"
		partialUpdateParams := domain.UpdateUserParams{
			ID:    createdUser.ID,
			Name:  partialUpdatedName,
			Email: updatedEmail, // Email remains unchanged
		}

		// Act - Partial update
		updatedUser, err := u.repository.UpdateUser(ctx, partialUpdateParams)

		// Assert - Ensure only the name is updated
		u.Require().NoError(err)
		u.Require().NotNil(updatedUser)
		u.Equal(partialUpdatedName, updatedUser.Name, "domain.User name should be updated")
		u.Equal(updatedEmail, updatedUser.Email, "domain.User email should remain unchanged")

		// Act - Retrieve and verify changes in the database
		retrievedUser, err := u.repository.GetUserByID(ctx, createdUser.ID)

		// Assert - Verify retrieved user matches updated details
		u.Require().NoError(err)
		u.Require().NotNil(retrievedUser)
		u.Equal(partialUpdatedName, retrievedUser.Name, "Retrieved user name should match the updated name")
		u.Equal(updatedEmail, retrievedUser.Email, "Retrieved user email should remain unchanged")
	})
}

func (u *UserTestSuite) TestUpdateUserRole() {
	ctx := u.ctx
	t := u.T() // Get the testing instance

	// Arrange - Create a sample user
	users, err := u.CreateSampleUsers(ctx, 1)
	u.Require().NoError(err)
	createdUser := users[0]

	t.Run("Update User Role to Admin", func(t *testing.T) {
		// Initial role should be "user"
		u.Equal("user", createdUser.Role, "Initial role should be 'user'")

		// Prepare update parameters
		updateParams := domain.UpdateUserRoleParams{
			ID:   createdUser.ID,
			Role: "admin",
		}

		// Act - Update the user's role
		updatedUser, err := u.repository.UpdateUserRole(ctx, updateParams)

		// Assert - Verify update
		u.Require().NoError(err, "Failed to update user role")
		u.Require().NotNil(updatedUser)
		u.Equal(createdUser.ID, updatedUser.ID, "User ID should remain unchanged")
		u.Equal("admin", updatedUser.Role, "User role should be updated to 'admin'")
		u.NotEqual(createdUser.UpdatedAt, updatedUser.UpdatedAt, "UpdatedAt should be changed")

		// Act - Retrieve the updated user to confirm changes in the database
		retrievedUser, err := u.repository.GetUserByID(ctx, createdUser.ID)

		// Assert - Verify retrieved user has the updated role
		u.Require().NoError(err, "Failed to retrieve updated user")
		u.Require().NotNil(retrievedUser)
		u.Equal("admin", retrievedUser.Role, "Retrieved user role should be 'admin'")
	})

	t.Run("Update User Role Back to User", func(t *testing.T) {
		// Prepare update parameters to change role back to "user"
		updateParams := domain.UpdateUserRoleParams{
			ID:   createdUser.ID,
			Role: "user",
		}

		// Act - Update the user's role
		updatedUser, err := u.repository.UpdateUserRole(ctx, updateParams)

		// Assert - Verify update
		u.Require().NoError(err, "Failed to update user role")
		u.Require().NotNil(updatedUser)
		u.Equal("user", updatedUser.Role, "User role should be updated to 'user'")

		// Act - Retrieve the updated user to confirm changes in the database
		retrievedUser, err := u.repository.GetUserByID(ctx, createdUser.ID)

		// Assert - Verify retrieved user has the updated role
		u.Require().NoError(err, "Failed to retrieve updated user")
		u.Require().NotNil(retrievedUser)
		u.Equal("user", retrievedUser.Role, "Retrieved user role should be 'user'")
	})

	t.Run("Update Non-Existent User", func(t *testing.T) {
		// Prepare update parameters with non-existent user ID
		nonExistentID := uuid.New()
		updateParams := domain.UpdateUserRoleParams{
			ID:   nonExistentID,
			Role: "admin",
		}

		// Act - Try to update non-existent user
		_, err := u.repository.UpdateUserRole(ctx, updateParams)

		// Assert - Should return ErrNotFound
		u.Require().Error(err, "Should error when updating non-existent user")
		u.ErrorIs(err, common.ErrNotFound, "Should return ErrNotFound")
	})
}

func (u *UserTestSuite) TestDeleteUser() {
	ctx := u.ctx
	t := u.T() // Get the testing instance

	t.Run("Delete Existing domain.User", func(t *testing.T) {
		// Arrange - Create a sample user
		users, err := u.CreateSampleUsers(ctx, 1)
		u.Require().NoError(err)
		createdUser := users[0]

		// Act - Delete the user
		err = u.repository.DeleteUser(ctx, createdUser.ID)

		// Assert - Verify deletion
		u.Require().NoError(err, "Failed to delete user")

		// Act - Try retrieving the deleted user
		_, err = u.repository.GetUserByID(ctx, createdUser.ID)

		// Assert - domain.User should no longer exist
		u.Require().Error(err)
		u.ErrorIs(err, common.ErrNotFound, "not found")
	})

	t.Run("Delete Non-Existent domain.User", func(t *testing.T) {
		// Arrange - Generate a random UUID
		nonExistentID := uuid.New()

		// Act - Try deleting a user that doesn't exist
		err := u.repository.DeleteUser(ctx, nonExistentID)

		// Assert - Should return ErrNotFound
		u.Require().Error(err)
		u.ErrorIs(err, common.ErrNotFound, "Expected ErrNotFound for non-existent user")
	})
}
