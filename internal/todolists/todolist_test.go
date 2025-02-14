//go:build unit

package todolist

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
	_, err := t.pgt.DB().Exec(t.ctx, "TRUNCATE TABLE todolists CASCADE;")
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

func (t *TodoListTestSuite) setupTodoLists(ctx context.Context, userID uuid.UUID, count int) ([]TodoList, error) {
	var createdLists []TodoList

	for i := 1; i <= count; i++ {
		params := CreateTodoListParams{
			UserID:      userID,
			Title:       fmt.Sprintf("Todo List %d", i),
			Description: fmt.Sprintf("Description %d", i),
		}

		todoList, err := t.store.CreateTodoList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("failed to create todo list %d: %w", i, err)
		}

		createdLists = append(createdLists, todoList)
	}

	return createdLists, nil
}

// TestCreateTodoList validates the creation of a new todo list
func (t *TodoListTestSuite) TestCreateTodoList() {
	ctx := t.ctx

	// Arrange
	params := CreateTodoListParams{
		UserID:      t.userID, // Use the user created in SetupTest
		Title:       "My Todo List",
		Description: "A test description",
	}

	// Act
	createdTodoList, err := t.store.CreateTodoList(ctx, params)

	// Assert
	t.Require().NoError(err)
	t.Require().NotNil(createdTodoList)

	// Verify the created record matches the input
	t.Equal(params.Title, createdTodoList.Title)
	t.Equal(params.Description, createdTodoList.Description)
	t.Equal(params.UserID, createdTodoList.UserID)

	// Ensure generated fields are valid
	t.Require().NotEqual(uuid.Nil, createdTodoList.ID)
	t.Require().NotZero(createdTodoList.CreatedAt)
	t.Require().NotZero(createdTodoList.UpdatedAt)
}

func (t *TodoListTestSuite) TestUpdateTodoList() {
	ctx := t.ctx
	userID := t.userID

	// Arrange: Create a single todo list using the setup function
	createdLists, err := t.setupTodoLists(ctx, userID, 1)
	t.Require().NoError(err)
	t.Require().Len(createdLists, 1)

	createdTodoList := createdLists[0] // Get the first (and only) todo list

	// Define new values for the update
	updateParams := UpdateTodoListParams{
		ID:          createdTodoList.ID,
		UserID:      userID,
		Title:       "Updated Todo List",
		Description: "Updated Description",
	}

	// Act: Update the todo list and get the updated result directly
	updatedTodoList, err := t.store.UpdateTodoList(ctx, updateParams)

	// Assert: Verify the updated todo list
	t.Require().NoError(err)
	t.Require().NotNil(updatedTodoList)
	t.Equal(createdTodoList.ID, updatedTodoList.ID)
	t.Equal(userID, updatedTodoList.UserID)
	t.Equal(updateParams.Title, updatedTodoList.Title)
	t.Equal(updateParams.Description, updatedTodoList.Description)
}

func (t *TodoListTestSuite) TestListTodoListsWithPagination() {
	ctx := t.ctx
	userID := t.userID

	// Arrange: Create 5 todo lists using the setup function
	totalLists := 5
	createdLists, err := t.setupTodoLists(ctx, userID, totalLists)
	t.Require().NoError(err)
	t.Require().Len(createdLists, totalLists)

	// Reverse the created lists to match expected descending order
	reversedLists := make([]TodoList, len(createdLists))
	for i, list := range createdLists {
		reversedLists[len(createdLists)-1-i] = list
	}

	// Act: Retrieve paginated results
	limit := int32(2)
	offset := int32(0)
	params := ListTodoListsWithPaginationParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	results, err := t.store.ListTodoListsWithPagination(ctx, params)

	// Assert: Verify correct retrieval
	t.Require().NoError(err)
	t.Require().NotNil(results)
	t.Require().Len(results, int(limit))

	// Verify contents of the retrieved todo lists
	expected := reversedLists[:limit]
	for i, result := range results {
		t.Equal(expected[i].Title, result.Title)
		t.Equal(expected[i].Description, result.Description)
		t.Equal(userID, result.UserID)
	}

	// Act: Retrieve the next page
	offset = limit
	params.Offset = offset

	results, err = t.store.ListTodoListsWithPagination(ctx, params)

	// Assert: Verify retrieval of the next page
	t.Require().NoError(err)
	t.Require().NotNil(results)
	t.Require().Len(results, int(limit))

	expected = reversedLists[offset : offset+int32(limit)]
	for i, result := range results {
		t.Equal(expected[i].Title, result.Title)
		t.Equal(expected[i].Description, result.Description)
		t.Equal(userID, result.UserID)
	}
}

func (t *TodoListTestSuite) TestPartialBulkDeleteTodoLists() {
	ctx := t.ctx
	userID := t.userID

	// Arrange: Create 3 todo lists using setup function
	totalLists := 3
	createdLists, err := t.setupTodoLists(ctx, userID, totalLists)
	t.Require().NoError(err)
	t.Require().Len(createdLists, totalLists)

	// Collect all todo list IDs
	var todoListIDs []uuid.UUID
	for _, todo := range createdLists {
		todoListIDs = append(todoListIDs, todo.ID)
	}

	// Select only 2 out of 3 to delete
	idsToDelete := todoListIDs[:2] // Select the first two todo list IDs

	// Act: Bulk delete the selected todo lists
	deleteParams := DeleteTodoListsParams{
		UserID: userID,
		IDs:    idsToDelete,
	}
	rowsAffected, err := t.store.DeleteTodoLists(ctx, deleteParams)

	// Assert
	t.Require().NoError(err)
	t.Equal(int64(len(idsToDelete)), rowsAffected, "Expected rows affected to match number of todo lists deleted")

	// Verify the selected todo lists are deleted
	for _, id := range idsToDelete {
		_, err := t.store.GetTodoListByID(ctx, GetTodoListByIDParams{ID: id, UserID: userID})
		t.ErrorIs(err, common.ErrNotFound)
	}

	// Verify the remaining todo list still exists
	remainingID := todoListIDs[2]
	retrievedTodoList, err := t.store.GetTodoListByID(ctx, GetTodoListByIDParams{ID: remainingID, UserID: userID})
	t.Require().NoError(err)
	t.Require().NotNil(retrievedTodoList)
	t.Equal(createdLists[2].Title, retrievedTodoList.Title)
	t.Equal(createdLists[2].Description, retrievedTodoList.Description)
}
