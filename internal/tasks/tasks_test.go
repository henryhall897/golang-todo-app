//go:build unit

package tasks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"golang-todo-app/internal/core/common"
	"golang-todo-app/internal/core/dbpool"
	"golang-todo-app/pkg/dbtest"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TaskTestSuite struct {
	suite.Suite
	pgt        *dbtest.PostgresTest
	ctx        context.Context
	store      *Store
	userID     uuid.UUID // User ID for the tests
	todoListID uuid.UUID // Todo list ID for the tests
	taskID     uuid.UUID // Task ID for the tests
}

func TestTasks(t *testing.T) {
	suite.Run(t, &TaskTestSuite{})
}

func (t *TaskTestSuite) SetupSuite() {
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

	// Initialize the Store with both TodoListStore and UserStore
	t.store = New(t.pgt.DB())
}

func (t *TaskTestSuite) SetupTest() {
	// Create a user
	var err error
	t.userID, err = t.createUserDirect("Test User", "test@example.com")
	t.Require().NoError(err)

	// Create a todo list
	t.todoListID, err = t.createTodoListDirect(t.userID, "Sample Todo List", "This is a sample todo list")
	t.Require().NoError(err)
}

func (t *TaskTestSuite) TearDownSuite() {
	t.Require().NoError(t.pgt.TearDown())
}

func (t *TaskTestSuite) TearDownTest() {
	// Truncate the tasks table after each test
	_, err := t.pgt.DB().Exec(t.ctx, "TRUNCATE TABLE tasks CASCADE;")
	t.Require().NoError(err)

	// Truncate the todo_lists table after each test
	_, err = t.pgt.DB().Exec(t.ctx, "TRUNCATE TABLE todo_lists CASCADE;")
	t.Require().NoError(err)

	// Truncate the users table after each test
	_, err = t.pgt.DB().Exec(t.ctx, "TRUNCATE TABLE users CASCADE;")
	t.Require().NoError(err)
}

func (t *TaskTestSuite) createUserDirect(name, email string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := t.pgt.DB().QueryRow(
		t.ctx,
		"INSERT INTO users (id, name, email) VALUES (gen_random_uuid(), $1, $2) RETURNING id",
		name, email,
	).Scan(&userID)

	// Handle unique violation error
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return uuid.Nil, fmt.Errorf("user with email %s already exists", email)
		}
	}
	return userID, err
}

func (t *TaskTestSuite) createTodoListDirect(userID uuid.UUID, name, description string) (uuid.UUID, error) {
	var listID uuid.UUID
	err := t.pgt.DB().QueryRow(
		t.ctx,
		"INSERT INTO todo_lists (id, user_id, name, todo_desc) VALUES (gen_random_uuid(), $1, $2, $3) RETURNING id",
		userID, name, common.Ptr(description),
	).Scan(&listID)

	// Handle foreign key violation error
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			return uuid.Nil, fmt.Errorf("user with id %s does not exist", userID)
		}
	}
	return listID, err
}

func (t *TaskTestSuite) createMultipleSampleTasks(n int) ([]FullTask, error) {
	var tasks []FullTask

	for i := 1; i <= n; i++ {
		title := fmt.Sprintf("Task %d", i)
		description := fmt.Sprintf("This is task number %d", i)
		dueDate := time.Now().Add(time.Duration(i) * 24 * time.Hour)
		priority := int32(i) // Assign multiples of 10 as priority

		task, err := t.createSampleTask(title, description, "pending", dueDate, priority)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (t *TaskTestSuite) createSampleTask(title, description, status string, dueDate time.Time, priority int32) (FullTask, error) {
	return t.store.CreateTask(
		t.ctx,
		t.todoListID,            // List ID from the test suite
		title,                   // Title
		common.Ptr(description), // Description (pointer)
		common.Ptr(status),      // Status (pointer)
		dueDate,                 // Due date
		priority,                // Priority
	)
}

func extractTaskIDs(tasks []FullTask) []uuid.UUID {
	var ids []uuid.UUID
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	return ids
}

func (t *TaskTestSuite) TestCreateTask() {
	// Arrange: Use the bulk helper to create one task
	tasks, err := t.createMultipleSampleTasks(1)
	t.Require().NoError(err)
	t.Require().Len(tasks, 1)

	// Get the created task
	task := tasks[0]

	// Assert: Verify the task's values
	t.Require().NotNil(task.Title)
	t.Equal("Task 1", *task.Title)
	t.Require().NotNil(task.TaskDesc)
	t.Equal("This is task number 1", *task.TaskDesc)
	t.Require().NotNil(task.Status)
	t.Equal("pending", *task.Status)
	t.Require().NotNil(task.DueDate)
	t.WithinDuration(time.Now().Add(24*time.Hour), *task.DueDate, time.Second)
	t.Require().NotNil(task.Priority)
	t.Equal(int32(1), *task.Priority)
}

func (t *TaskTestSuite) TestUpdateTask() {
	// Arrange: Use the bulk helper to create one task
	tasks, err := t.createMultipleSampleTasks(1)
	t.Require().NoError(err)
	t.Require().Len(tasks, 1)

	// Get the created task
	createdTask := tasks[0]

	// New values for the update
	updatedTitle := "Updated Task Title"
	updatedDescription := "Updated description"
	updatedStatus := "completed"
	updatedDueDate := time.Now().Add(48 * time.Hour)
	updatedPriority := int32(2)

	// Act: Update the task
	params := UpdateTaskParams{
		ID:       createdTask.ID,
		ListID:   t.todoListID,
		UserID:   t.userID, // Pass UserID
		Title:    common.Ptr(updatedTitle),
		TaskDesc: common.Ptr(updatedDescription),
		Status:   common.Ptr(updatedStatus),
		DueDate:  common.Ptr(updatedDueDate),
		Priority: &updatedPriority,
	}
	updatedTask, err := t.store.UpdateTask(t.ctx, params)

	// Assert: Verify the updated values
	t.Require().NoError(err)
	t.Require().NotNil(updatedTask)
	t.Equal(updatedTitle, *updatedTask.Title)
	t.Equal(updatedDescription, *updatedTask.TaskDesc)
	t.Equal(updatedStatus, *updatedTask.Status)
	t.WithinDuration(updatedDueDate, *updatedTask.DueDate, time.Second)

	// ✅ Updated assertion for Priority
	t.Require().NotNil(updatedTask.Priority)
	t.Equal(updatedPriority, *updatedTask.Priority)
}

func (t *TaskTestSuite) TestUpdateTaskPriority() {
	// Arrange: Use the bulk helper to create one task
	tasks, err := t.createMultipleSampleTasks(10)
	t.Require().NoError(err)
	t.Require().Len(tasks, 10)

	// Get the created task
	createdTask := tasks[0]

	// New priority value
	newPriority := int32(5)

	// Act: Update the task's priority using the general UpdateTask function
	params := UpdateTaskParams{
		ID:       createdTask.ID,
		ListID:   t.todoListID,
		UserID:   t.userID,
		Priority: &newPriority,
	}
	updatedTask, err := t.store.UpdateTask(t.ctx, params)

	// Assert: Verify the updated values
	t.Require().NoError(err)
	t.Require().NotNil(updatedTask)
	t.Equal(createdTask.ID, updatedTask.ID)
	t.Require().NotNil(updatedTask.Priority)
	t.Equal(newPriority, *updatedTask.Priority)
	t.WithinDuration(time.Now(), *updatedTask.UpdatedAt, time.Second)
}

func (t *TaskTestSuite) TestMarkTaskCompleted() {
	// Arrange: Use the bulk helper to create one task
	tasks, err := t.createMultipleSampleTasks(1)
	t.Require().NoError(err)
	t.Require().Len(tasks, 1)

	// Get the created task
	createdTask := tasks[0]

	// New values for the update
	updatedStatus := "completed"

	// Act: Use the general UpdateTask function to mark the task as completed
	params := UpdateTaskParams{
		ID:          createdTask.ID,
		ListID:      t.todoListID,
		UserID:      t.userID,
		Status:      common.Ptr(updatedStatus), // Status should be set to 'completed'
		Priority:    nil,                       // Explicitly mark priority as NULL
		CompletedAt: common.Ptr(time.Now()),    // Set completed_at to now
	}
	updatedTask, err := t.store.UpdateTask(t.ctx, params)
	t.Require().NoError(err)
	t.Require().NotNil(updatedTask)

	// Assert: Verify that only the necessary fields are updated
	t.Equal("completed", *updatedTask.Status)                           // Ensure status is set to 'completed'
	t.Nil(updatedTask.Priority)                                         // Priority should be NULL
	t.NotNil(updatedTask.CompletedAt)                                   // Ensure completed_at is set
	t.WithinDuration(time.Now(), *updatedTask.CompletedAt, time.Second) // Ensure completed_at is within a second
	t.WithinDuration(time.Now(), *updatedTask.UpdatedAt, time.Second)   // Ensure updated_at is within a second

	// Verify other fields remain unchanged
	t.Equal(createdTask.ID, updatedTask.ID)
	t.Equal(createdTask.ListID, updatedTask.ListID)
	t.Equal(createdTask.Title, updatedTask.Title)
	t.Equal(createdTask.TaskDesc, updatedTask.TaskDesc)
	t.Equal(createdTask.DueDate, updatedTask.DueDate)
}

func (t *TaskTestSuite) TestDeleteTasks() {
	// Arrange: Create multiple sample tasks
	tasks, err := t.createMultipleSampleTasks(3)
	t.Require().NoError(err)
	t.Require().Len(tasks, 3)

	// Collect task IDs to delete
	taskIDs := extractTaskIDs(tasks)

	// Act: Delete the tasks
	params := DeleteTasksParams{
		IDs:    taskIDs,
		ListID: t.todoListID,
		UserID: t.userID,
	}
	deletedTasks, err := t.store.DeleteTasks(t.ctx, params)

	// Assert: Verify the correct number of tasks were deleted
	t.Require().NoError(err)
	t.Require().Len(deletedTasks, 3)

	// Verify that the deleted tasks match the input IDs
	for _, deletedTask := range deletedTasks {
		t.Contains(taskIDs, deletedTask.ID)
		t.Equal(t.todoListID, deletedTask.ListID)
		t.Equal("pending", *deletedTask.Status) // Original status
	}
}

func (t *TaskTestSuite) TestListTasks() {
	// Arrange: Create multiple tasks using the helper
	tasks, err := t.createMultipleSampleTasks(5)
	t.Require().NoError(err)
	t.Require().Len(tasks, 5)

	// Prepare the ListTasksParams
	params := TaskListParams{
		ListID: t.todoListID,
		UserID: t.userID,
	}

	// Act: Call the ListTasks function
	listedTasks, err := t.store.ListTasks(t.ctx, params)

	// Assert
	t.Require().NoError(err)
	t.Require().NotNil(listedTasks)
	t.Require().Len(listedTasks, 5)

	// Verify the tasks are listed in the correct order
	for i, task := range listedTasks {
		t.Equal(tasks[i].ID, task.ID)
		t.Equal(tasks[i].Title, task.Title)
		t.Equal(tasks[i].ListID, task.ListID)
		t.Equal(*tasks[i].TaskDesc, *task.TaskDesc)
		t.Equal(*tasks[i].Status, *task.Status)
		t.WithinDuration(*tasks[i].DueDate, *task.DueDate, time.Second)
		t.Equal(tasks[i].Priority, task.Priority)
	}
}

func (t *TaskTestSuite) TestListOverdueTasks() {
	// Arrange: Create sample tasks, some overdue and some not
	tasks, err := t.createMultipleSampleTasks(5)
	t.Require().NoError(err)
	t.Require().Len(tasks, 5)

	// Mark the first three tasks as overdue
	overdueTasks := tasks[:3]
	for i, task := range overdueTasks {
		overdueDueDate := time.Now().Add(-time.Duration(i+1) * 24 * time.Hour) // Past dates

		// Update the task using UpdateTask to set the overdue due_date
		_, err := t.store.UpdateTask(t.ctx, UpdateTaskParams{
			ID:      task.ID,
			ListID:  t.todoListID,
			UserID:  t.userID,
			DueDate: common.Ptr(overdueDueDate),
			Status:  common.Ptr("pending"),
		})
		t.Require().NoError(err)
	}

	// Act: List overdue tasks
	params := TaskListParams{
		ListID: t.todoListID,
		UserID: t.userID,
	}
	result, err := t.store.ListOverdueTasks(t.ctx, params)

	// Assert
	t.Require().NoError(err)
	t.Require().NotNil(result)
	t.Len(result, len(overdueTasks))

	// Verify that all returned tasks are overdue
	for _, task := range result {
		t.True(task.DueDate.Before(time.Now()))
		t.NotEqual("completed", *task.Status)
	}
}

func (t *TaskTestSuite) TestListTasksByStatus() {
	// Arrange: Use the bulk helper to create tasks with different statuses
	tasks, err := t.createMultipleSampleTasks(5)
	t.Require().NoError(err)
	t.Require().Len(tasks, 5)

	// Assign a specific status for the test
	statusToFilter := "pending"

	// Act: Retrieve tasks with the specified status
	params := CountTasksByStatusParams{
		ListID: t.todoListID,
		UserID: t.userID,
		Status: &statusToFilter, // directly pass the status string
	}
	result, err := t.store.ListTasksByStatus(t.ctx, params)

	// Assert: Verify the returned tasks match the expected status
	t.Require().NoError(err)
	t.Require().NotNil(result)

	// Verify that all tasks have the correct status
	for _, task := range result {
		t.Equal(statusToFilter, *task.Status)
	}

	// Verify other fields are properly returned
	t.Require().Len(result, len(tasks)) // Make sure the number of tasks returned is correct
	for i, task := range result {
		t.Equal(tasks[i].ID, task.ID)           // Ensure IDs match
		t.Equal(tasks[i].Title, task.Title)     // Ensure Titles match
		t.Equal(tasks[i].DueDate, task.DueDate) // Ensure DueDate matches

		// Handle time comparison by dereferencing the pointers to time values
		// If tasks[i].UpdatedAt is a *time.Time, we need to dereference it in the comparison
		t.WithinDuration(*tasks[i].UpdatedAt, *task.UpdatedAt, time.Second) // Compare UpdatedAt properly
		t.WithinDuration(*tasks[i].CreatedAt, *task.CreatedAt, time.Second) // Compare CreatedAt properly
	}
}

func (t *TaskTestSuite) TestUpdateSomeTasksToCompletedAndSearchCompleted() {
	// Arrange: Create 10 tasks, some will be marked as completed
	tasks, err := t.createMultipleSampleTasks(10)
	t.Require().NoError(err)
	t.Require().Len(tasks, 10)

	// Mark the first 4 tasks as completed
	for i := 0; i < 4; i++ {
		updatedStatus := "completed"
		params := UpdateTaskParams{
			ID:          tasks[i].ID,
			ListID:      t.todoListID,
			UserID:      t.userID,
			Status:      common.Ptr(updatedStatus),
			Priority:    nil,                    // Explicitly mark priority as NULL
			CompletedAt: common.Ptr(time.Now()), // Set completed_at to now
		}
		_, err := t.store.UpdateTask(t.ctx, params)
		t.Require().NoError(err)
	}

	// Act: Retrieve tasks with the "completed" status
	statusToFilter := "completed"
	params := CountTasksByStatusParams{
		ListID: t.todoListID,
		UserID: t.userID,
		Status: &statusToFilter, // Pass the address of the status string
	}
	completedTasks, err := t.store.ListTasksByStatus(t.ctx, params)

	// Assert: Verify the returned tasks match the expected status
	t.Require().NoError(err)
	t.Require().NotNil(completedTasks)
	t.Require().Len(completedTasks, 4) // Ensure 4 tasks are completed

	// Verify that all returned tasks have the "completed" status
	for _, task := range completedTasks {
		t.Equal("completed", *task.Status)
		t.Nil(task.Priority)                                         // This should now be NULL for completed tasks
		t.NotNil(task.CompletedAt)                                   // Ensure CompletedAt is set
		t.WithinDuration(time.Now(), *task.CompletedAt, time.Second) // Verify the CompletedAt is recent
	}

	// Verify that the tasks with non-completed status are unaffected
	nonCompletedTasks := tasks[4:]
	for _, task := range nonCompletedTasks {
		t.NotEqual("completed", *task.Status) // Ensure non-completed tasks are not returned
	}
}

func (t *TaskTestSuite) TestSearchTasks() {
	// Arrange: Create sample tasks with different titles/descriptions
	tasks, err := t.createMultipleSampleTasks(5)
	t.Require().NoError(err)
	t.Require().Len(tasks, 5)

	// Set the keyword for searching
	keyword := common.Ptr("Task 1")

	// Act: Search for tasks matching the keyword
	params := SearchTasksParams{
		ListID:  t.todoListID,
		UserID:  t.userID,
		Keyword: keyword,
	}
	result, err := t.store.SearchTasks(t.ctx, params)

	// Assert: Verify the results
	t.Require().NoError(err)
	t.Require().NotNil(result)

	// Ensure that task.Title is not nil and then dereference it
	for _, task := range result {
		t.Require().NotNil(task.Title) // Ensure Title is not nil
		// Check if the dereferenced Title contains the keyword
		t.Contains(*task.Title, *keyword) // Dereference Title and check if it contains keyword
	}
}

func (t *TaskTestSuite) TestUpdateMultipleTaskPriorities() {
	// Arrange: Create 5 tasks with different priorities (multiples of 10)
	tasks, err := t.createMultipleSampleTasks(5)
	t.Require().NoError(err)
	t.Require().Len(tasks, 5)

	// Select one of the tasks to update its priority (let's change the priority of Task 3)
	updatedPriority := common.Ptr(int32(5)) // Priority value to be updated (lower value, higher priority)

	// Get the task that we want to update
	taskToUpdate := tasks[2] // Task 3 in the list

	// Act: Update the priority of Task 3 using the UpdateTaskPriority function
	params := UpdateTaskParams{
		ID:       taskToUpdate.ID,
		ListID:   t.todoListID,
		UserID:   t.userID,
		Priority: updatedPriority, // Updated priority pointer
	}
	updatedTask, err := t.store.UpdateTask(t.ctx, params)
	t.Require().NoError(err)
	t.Require().NotNil(updatedTask)

	// Assert: Verify that the updated priority is correct
	t.Equal(updatedPriority, updatedTask.Priority)

	// Act: Retrieve all tasks again after the priority update
	allTasks, err := t.store.ListTasks(t.ctx, TaskListParams{
		ListID: t.todoListID,
		UserID: t.userID,
	})
	t.Require().NoError(err)
	t.Require().Len(allTasks, 5)

	// Assert: Verify that tasks are ordered correctly (ascending order of priority)
	for i := 1; i < len(allTasks); i++ {
		t.Assert().LessOrEqual(*allTasks[i-1].Priority, *allTasks[i].Priority)
	}

	// Verify other fields remain unchanged for the task we updated
	t.Equal(taskToUpdate.Title, updatedTask.Title)
	t.Equal(taskToUpdate.TaskDesc, updatedTask.TaskDesc)
	t.Equal(taskToUpdate.Status, updatedTask.Status)
	t.Equal(taskToUpdate.DueDate, updatedTask.DueDate)
}
