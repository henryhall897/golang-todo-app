//go:build unit

package todolist

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

type TodoListTestSuite struct {
	suite.Suite
	pgt    *dbtest.PostgresTest
	ctx    context.Context
	store  *Store
	userID uuid.UUID // Store a user ID that all tests can use
}

func TestTodoLists(t *testing.T) {
	suite.Run(t, &TodoListTestSuite{})
}

func (t *TodoListTestSuite) SetupSuite() {
	t.ctx = context.Background()

	var err error
	t.pgt, err = dbtest.NewPostgresTest(t.ctx, zap.L(), "../../migrations", &dbpool.Config{
		Logging:      false,
		Host:         "localhost",
		Port:         "5432",
		User:         "testuser",
		Password:     "1234",
		DatabaseName: "todotestdb",
		MaxConns:     1,
		MinConns:     1,
	})
	t.Require().NoError(err)

	err = t.pgt.MigrateUp()
	t.Require().NoError(err)

	t.store = New(t.pgt.DB())
}

func (t *TodoListTestSuite) SetupTest() {
	// Create a valid user before each test
	var err error
	t.userID, err = t.createUser("Test User", "test@example.com")
	t.Require().NoError(err)
}

func (t *TodoListTestSuite) TearDownSuite() {
	t.Require().NoError(t.pgt.TearDown())
}

func (t *TodoListTestSuite) TearDownTest() {
	// Truncate the todo_lists table after each test
	_, err := t.pgt.DB().Exec(t.ctx, "TRUNCATE TABLE todo_lists CASCADE;")
	t.Require().NoError(err)

	// Truncate the users table after each test
	_, err = t.pgt.DB().Exec(t.ctx, "TRUNCATE TABLE users CASCADE;")
	t.Require().NoError(err)
}

func (t *TodoListTestSuite) createUser(name, email string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := t.pgt.DB().QueryRow(t.ctx, "INSERT INTO users (id, name, email) VALUES (gen_random_uuid(), $1, $2) RETURNING id", name, email).Scan(&userID)
	return userID, err
}

// TestCreateTodoList validates the creation of a new todo list
func (t *TodoListTestSuite) TestCreateTodoList() {
	ctx := t.ctx

	// Arrange
	userID := t.userID // Use the user created in SetupTest
	name := "My Todo List"
	description := "A test description"

	// Act
	createdTodoList, err := t.store.CreateTodoList(ctx, userID, name, description)

	// Assert
	t.Require().NoError(err)
	t.Require().NotNil(createdTodoList)

	// Verify the created record matches the input
	t.Equal(name, createdTodoList.Name)
	t.Equal(description, createdTodoList.Description)
	t.Equal(userID, createdTodoList.UserID)

	// Ensure generated fields are valid
	t.Require().NotEqual(uuid.Nil, createdTodoList.ID)
	t.Require().NotZero(createdTodoList.CreatedAt)
	t.Require().NotZero(createdTodoList.UpdatedAt)
}

func (t *TodoListTestSuite) TestUpdateTodoList() {
	ctx := t.ctx

	// Arrange: Create a user and a todo list
	userID := t.userID
	name := "Initial Todo List"
	description := "Initial Description"

	// Create the initial todo list
	createdTodoList, err := t.store.CreateTodoList(ctx, userID, name, description)
	t.Require().NoError(err)

	// New values for the update
	updatedName := "Updated Todo List"
	updatedDescription := "Updated Description"

	// Act: Update the todo list and get the updated result directly
	updatedTodoList, err := t.store.UpdateTodoList(ctx, createdTodoList.ID, userID, updatedName, updatedDescription)

	// Assert: Verify the updated todo list
	t.Require().NoError(err)
	t.Require().NotNil(updatedTodoList)
	t.Equal(createdTodoList.ID, updatedTodoList.ID)
	t.Equal(userID, updatedTodoList.UserID)
	t.Equal(updatedName, updatedTodoList.Name)
	t.Equal(updatedDescription, updatedTodoList.Description)

	// Since the `UpdateTodoList` function now returns the updated row, no need to fetch it again from the database
}

func (t *TodoListTestSuite) TestListTodoListsWithPagination() {
	ctx := t.ctx

	// Arrange: Create a user and several todo lists
	userID := t.userID
	todoLists := []struct {
		Name        string
		Description string
	}{
		{"Todo List 1", "Description 1"},
		{"Todo List 2", "Description 2"},
		{"Todo List 3", "Description 3"},
	}

	// Insert the todo lists
	for _, todo := range todoLists {
		_, err := t.store.CreateTodoList(ctx, userID, todo.Name, todo.Description)
		t.Require().NoError(err)
	}

	// Reverse the `todoLists` slice to reflect DESC order
	reversedTodoLists := []struct {
		Name        string
		Description string
	}{
		{"Todo List 3", "Description 3"},
		{"Todo List 2", "Description 2"},
		{"Todo List 1", "Description 1"},
	}

	// Act: Retrieve the todo lists with pagination
	limit := int32(2)
	offset := int32(0)
	results, err := t.store.ListTodoListsWithPagination(ctx, userID, limit, offset)

	// Assert
	t.Require().NoError(err)
	t.Require().NotNil(results)
	t.Require().Len(results, int(limit)) // Verify the number of results matches the limit

	// Verify the contents of the retrieved todo lists
	expected := reversedTodoLists[:limit] // First 2 items in DESC order
	for i, result := range results {
		t.Equal(expected[i].Name, result.Name)
		t.Equal(expected[i].Description, result.Description)
		t.Equal(userID, result.UserID)
	}

	// Act: Retrieve the next page
	offset = limit
	results, err = t.store.ListTodoListsWithPagination(ctx, userID, limit, offset)

	// Assert
	t.Require().NoError(err)
	t.Require().NotNil(results)
	t.Require().Len(results, 1) // Only one item should remain

	// Verify the contents of the remaining todo list
	expected = reversedTodoLists[offset:]
	for i, result := range results {
		t.Equal(expected[i].Name, result.Name)
		t.Equal(expected[i].Description, result.Description)
		t.Equal(userID, result.UserID)
	}
}

func (t *TodoListTestSuite) TestDeleteTodoList() {
	ctx := t.ctx

	// Arrange: Use the user created in SetupTest
	userID := t.userID
	name := "Sample Todo List"
	description := "Sample Description"

	// Create a todo list for testing
	createdTodoList, err := t.store.CreateTodoList(ctx, userID, name, description)
	t.Require().NoError(err)

	// Act: Delete the todo list
	rowsAffected, err := t.store.DeleteTodoList(ctx, createdTodoList.ID, userID)

	// Assert
	t.Require().NoError(err)
	t.Equal(int64(1), rowsAffected, "Expected 1 row to be affected")

	// Verify the todo list is deleted
	_, err = t.store.GetTodoListByID(ctx, createdTodoList.ID, userID)
	t.ErrorIs(err, common.ErrNotFound)
}

func (t *TodoListTestSuite) TestBulkDeleteTodoLists() {
	ctx := t.ctx

	// Arrange
	userID := t.userID // Reuse the user ID created in SetupTest
	name := "Sample Todo List"
	description := "Sample Description"

	// Create todo lists for testing and collect their IDs
	var todoListIDs []uuid.UUID
	for i := 0; i < 3; i++ {
		createdTodoList, err := t.store.CreateTodoList(ctx, userID, name, description)
		t.Require().NoError(err)
		todoListIDs = append(todoListIDs, createdTodoList.ID)
	}

	// Act: Bulk delete the todo lists
	rowsAffected, err := t.store.BulkDeleteTodoLists(ctx, todoListIDs, userID)

	// Assert
	t.Require().NoError(err)
	t.Equal(int64(len(todoListIDs)), rowsAffected, "Expected rows affected to match number of todo lists deleted")

	// Verify the todo lists are deleted
	for _, id := range todoListIDs {
		_, err := t.store.GetTodoListByID(ctx, id, userID)
		t.ErrorIs(err, common.ErrNotFound)
	}
}

func (t *TodoListTestSuite) TestPartialBulkDeleteTodoLists() {
	ctx := t.ctx

	// Arrange
	userID := t.userID // Reuse the user ID created in SetupTest
	name := "Sample Todo List"
	description := "Sample Description"

	// Create todo lists for testing and collect their IDs
	var todoListIDs []uuid.UUID
	for i := 0; i < 3; i++ {
		createdTodoList, err := t.store.CreateTodoList(ctx, userID, name, description)
		t.Require().NoError(err)
		todoListIDs = append(todoListIDs, createdTodoList.ID)
	}
	// Select only 2 out of 3 to delete
	idsToDelete := todoListIDs[:2] // Select the first two todo list IDs

	// Act: Bulk delete the selected todo lists
	rowsAffected, err := t.store.BulkDeleteTodoLists(ctx, idsToDelete, userID)

	// Assert
	t.Require().NoError(err)
	t.Equal(int64(len(idsToDelete)), rowsAffected, "Expected rows affected to match number of todo lists deleted")

	// Verify the selected todo lists are deleted
	for _, id := range idsToDelete {
		_, err := t.store.GetTodoListByID(ctx, id, userID)
		t.ErrorIs(err, common.ErrNotFound)
	}

	// Verify the remaining todo list still exists
	remainingID := todoListIDs[2]
	retrievedTodoList, err := t.store.GetTodoListByID(ctx, remainingID, userID)
	t.Require().NoError(err)
	t.Require().NotNil(retrievedTodoList)
	t.Equal(name, retrievedTodoList.Name)
	t.Equal(description, retrievedTodoList.Description)
}
