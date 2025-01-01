//go:build unit

package users

import (
	"context"
	"testing"

	"golang-todo-app/internal/core/common"
	"golang-todo-app/internal/core/dbpool"
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

// // TestListUsers validates listing all users
// func (u *UserTestSuite) TestListUsers() {
// 	ctx := u.ctx

// 	// Arrange
// 	users := []struct {
// 		Name  string
// 		Email string
// 	}{
// 		{"Alice", "alice@example.com"},
// 		{"Bob", "bob@example.com"},
// 		{"Charlie", "charlie@example.com"},
// 	}

// 	for _, user := range users {
// 		_, err := u.store.CreateUser(ctx, user.Name, user.Email)
// 		u.Require().NoError(err)
// 	}

// 	// Act
// 	results, err := u.store.ListUsers(ctx)

// 	// Assert
// 	u.Require().NoError(err)
// 	u.Require().Len(results, len(users))

// 	for i, result := range results {
// 		u.Equal(users[i].Name, result.Name)
// 		u.Equal(users[i].Email, result.Email)
// 	}
// }
