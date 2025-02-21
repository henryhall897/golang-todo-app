package tasks

import (
	"testing"
	"time"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/tasks/gen"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestToDBCreateTask(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	validTime := time.Now()

	// Initialize CreateTaskParams
	originalTask := CreateTaskParams{
		ListID:      validUUID,
		Title:       common.Ptr("Sample Task Title"),
		Description: common.Ptr("This is a sample task description."),
		Status:      common.Ptr("pending"),
		DueDate:     &validTime,
		Priority:    1,
	}

	// Act
	dbTask, err := toDBCreateTask(originalTask)

	// Assert
	require.NoError(t, err)

	// Validate that the transformed fields match the original ones
	require.Equal(t, originalTask.ListID[:], dbTask.ListID.Bytes[:])
	require.Equal(t, *originalTask.Title, dbTask.Title.String)
	require.Equal(t, *originalTask.Description, dbTask.Description.String)
	require.Equal(t, *originalTask.Status, dbTask.Status.String)
	require.WithinDuration(t, *originalTask.DueDate, dbTask.DueDate.Time, time.Second)
	require.Equal(t, originalTask.Priority, dbTask.Priority.Int32)
}

func TestToFullTask(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	validTime := time.Now().UTC()

	// Create a gen.Task struct to use as input
	genTask := gen.Task{
		ID:          pgtype.UUID{Bytes: validUUID, Valid: true},
		ListID:      pgtype.UUID{Bytes: validUUID, Valid: true},
		Title:       pgtype.Text{String: "Sample Task Title", Valid: true},
		Description: pgtype.Text{String: "This is a sample task description.", Valid: true},
		Status:      pgtype.Text{String: "pending", Valid: true},
		DueDate:     pgtype.Timestamp{Time: validTime, Valid: true},
		CreatedAt:   pgtype.Timestamp{Time: validTime, Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: validTime, Valid: true},
		Priority:    pgtype.Int4{Int32: 5, Valid: true},
		CompletedAt: pgtype.Timestamp{Valid: false}, // CompletedAt is not set for this task
	}

	// Act: Transform gen.Task into FullTask
	fullTask, err := toFullTask(genTask)

	// Assert: Check that the transformation worked correctly
	require.NoError(t, err)
	require.Equal(t, validUUID, fullTask.ID)
	require.Equal(t, validUUID, fullTask.ListID)
	require.Equal(t, "Sample Task Title", *fullTask.Title)
	require.Equal(t, "This is a sample task description.", *fullTask.Description)
	require.Equal(t, "pending", *fullTask.Status)
	require.WithinDuration(t, validTime, *fullTask.DueDate, time.Second)
	require.WithinDuration(t, validTime, fullTask.CreatedAt, time.Second)
	require.WithinDuration(t, validTime, fullTask.UpdatedAt, time.Second)
	require.Equal(t, int32(5), *fullTask.Priority)
	require.Nil(t, fullTask.CompletedAt) // CompletedAt should be nil

	// Confirm if we transformed all fields correctly
	require.Equal(t, validUUID, fullTask.ID)
	require.Equal(t, validUUID, fullTask.ListID)
	require.Equal(t, "Sample Task Title", *fullTask.Title)
	require.Equal(t, "This is a sample task description.", *fullTask.Description)
	require.Equal(t, "pending", *fullTask.Status)
	require.Equal(t, validTime, *fullTask.DueDate)
	require.Equal(t, validTime, fullTask.CreatedAt)
	require.Equal(t, validTime, fullTask.UpdatedAt)
	require.Equal(t, int32(5), *fullTask.Priority)
	require.Nil(t, fullTask.CompletedAt)
}
func TestToFullTaskList(t *testing.T) {
	// Arrange: Create multiple gen.Task structs to use as input
	validUUID := uuid.New()
	validTime := time.Now()

	genTasks := []gen.Task{
		{
			ID:          pgtype.UUID{Bytes: validUUID, Valid: true},
			ListID:      pgtype.UUID{Bytes: validUUID, Valid: true},
			Title:       pgtype.Text{String: "Task 1", Valid: true},
			Description: pgtype.Text{String: "This is task number 1", Valid: true},
			Status:      pgtype.Text{String: "pending", Valid: true},
			DueDate:     pgtype.Timestamp{Time: validTime, Valid: true},
			CreatedAt:   pgtype.Timestamp{Time: validTime, Valid: true},
			UpdatedAt:   pgtype.Timestamp{Time: validTime, Valid: true},
			Priority:    pgtype.Int4{Int32: 1, Valid: true},
			CompletedAt: pgtype.Timestamp{Valid: false}, // CompletedAt is not set
		},
		{
			ID:          pgtype.UUID{Bytes: validUUID, Valid: true},
			ListID:      pgtype.UUID{Bytes: validUUID, Valid: true},
			Title:       pgtype.Text{String: "Task 2", Valid: true},
			Description: pgtype.Text{String: "This is task number 2", Valid: true},
			Status:      pgtype.Text{String: "pending", Valid: true},
			DueDate:     pgtype.Timestamp{Time: validTime, Valid: true},
			CreatedAt:   pgtype.Timestamp{Time: validTime, Valid: true},
			UpdatedAt:   pgtype.Timestamp{Time: validTime, Valid: true},
			Priority:    pgtype.Int4{Int32: 2, Valid: true},
			CompletedAt: pgtype.Timestamp{Valid: false}, // CompletedAt is not set
		},
	}

	// Act: Transform gen.Task slice into FullTask slice
	fullTasks, err := toFullTaskList(genTasks)

	// Assert: Verify the transformation worked correctly
	require.NoError(t, err)
	require.Len(t, fullTasks, len(genTasks)) // Ensure the length of both slices matches
	// Verify each FullTask field is correctly populated
	// Verify each FullTask field is correctly populated
	for i := range fullTasks {
		require.Equal(t, genTasks[i].ID.Bytes[:], fullTasks[i].ID[:])                              // Compare the bytes correctly
		require.Equal(t, genTasks[i].ListID.Bytes[:], fullTasks[i].ListID[:])                      // Compare ListID bytes
		require.Equal(t, genTasks[i].Title.String, *fullTasks[i].Title)                            // Compare Title
		require.Equal(t, genTasks[i].Description.String, *fullTasks[i].Description)                // Compare TaskDesc
		require.Equal(t, genTasks[i].Status.String, *fullTasks[i].Status)                          // Compare Status
		require.WithinDuration(t, genTasks[i].DueDate.Time, *fullTasks[i].DueDate, time.Second)    // Compare DueDate
		require.WithinDuration(t, genTasks[i].CreatedAt.Time, fullTasks[i].CreatedAt, time.Second) // Compare CreatedAt
		require.WithinDuration(t, genTasks[i].UpdatedAt.Time, fullTasks[i].UpdatedAt, time.Second) // Compare UpdatedAt
		require.Equal(t, genTasks[i].Priority.Int32, *fullTasks[i].Priority)                       // Compare Priority
		require.Nil(t, fullTasks[i].CompletedAt)                                                   // Ensure CompletedAt is nil
	}

}

func TestToDBUpdateTaskParams(t *testing.T) {
	// Arrange: Create an UpdateTaskParams with sample values
	taskID := uuid.New()
	listID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	priority := int32(3)
	title := "Sample Task Title"
	Description := "Sample Task Description"
	status := "pending"

	params := UpdateTaskParams{
		ID:          taskID,
		ListID:      listID,
		UserID:      userID,
		Title:       &title,
		Description: &Description,
		Status:      &status,
		DueDate:     &now,
		Priority:    &priority,
		CompletedAt: nil, // Not completed task
	}

	// Act: Transform to DB-compatible struct
	dbParams, err := toDBUpdateTaskParams(params)

	// Assert: Ensure no error occurred during transformation
	require.NoError(t, err)

	// Verify the transformation was done correctly
	require.Equal(t, taskID[:], dbParams.ID.Bytes[:])                        // Check task ID
	require.Equal(t, "Sample Task Title", dbParams.Title.String)             // Check title
	require.Equal(t, "Sample Task Description", dbParams.Description.String) // Check task description
	require.Equal(t, "pending", dbParams.Status.String)                      // Check status
	require.WithinDuration(t, now, dbParams.DueDate.Time, time.Second)       // Check due date
	require.Equal(t, priority, dbParams.Priority.Int32)                      // Check priority
	require.False(t, dbParams.CompletedAt.Valid)                             // Check that CompletedAt is not valid
}

func TestToMarkTaskCompletedParams(t *testing.T) {
	// Arrange: Create sample data for UpdateTaskParams with pgtype fields
	taskID := uuid.New()
	userID := uuid.New()

	// Initialize the gen.UpdateTaskParams with pgtype fields
	params := gen.UpdateTaskParams{
		ID:     pgtype.UUID{Bytes: taskID, Valid: true}, // pgtype.UUID for ID
		UserID: pgtype.UUID{Bytes: userID, Valid: true}, // pgtype.UUID for UserID
	}

	// Act: Call the function to transform UpdateTaskParams into MarkTaskCompletedParams
	result, err := toMarkTaskCompletedParams(params)

	// Assert: Ensure no error occurred during the transformation
	require.NoError(t, err)

	// Verify that the transformation is correct
	require.True(t, result.ID.Valid)       // Verify that ID is valid
	require.Equal(t, params.ID, result.ID) // Verify that the ID bytes match

	require.True(t, result.UserID.Valid)           // Verify that UserID is valid
	require.Equal(t, params.UserID, result.UserID) // Verify that the UserID bytes match
}

// TestToDBDeleteTasksParams tests the toDBDeleteTasksParams function
func TestToDBDeleteTasksParams(t *testing.T) {
	// Arrange: Create sample input data
	taskID1 := uuid.New()
	taskID2 := uuid.New()
	userID := uuid.New()

	// Create the DeleteTasksParams struct
	params := DeleteTasksParams{
		IDs:    []uuid.UUID{taskID1, taskID2}, // Two tasks to delete
		ListID: uuid.New(),                    // Random Todo List ID
		UserID: userID,                        // User ID
	}

	// Act: Call the function to transform DeleteTasksParams into gen.DeleteTasksParams
	result, err := toDBDeleteTasksParams(params)

	// Assert: Ensure no error occurred during the transformation
	require.NoError(t, err)

	// Verify that the UserID field was correctly transformed to pgtype.UUID
	require.True(t, result.UserID.Valid)
	require.Equal(t, userID[:], result.UserID.Bytes[:])

	// Verify that the IDs field was correctly transformed to pgtype.UUID for each task
	require.Len(t, result.Column1, len(params.IDs))
	for i, dbTaskID := range result.Column1 {
		require.True(t, dbTaskID.Valid)
		require.Equal(t, params.IDs[i][:], dbTaskID.Bytes[:])
	}
}

// TestToDBListTasksParams tests the toDBListTasksParams function
func TestToDBListTasksParams(t *testing.T) {
	// Arrange: Create sample input data
	listID := uuid.New()
	userID := uuid.New()

	// Create the TaskListParams struct
	params := TaskListParams{
		ListID: listID, // Todo List ID
		UserID: userID, // User ID
	}

	// Act: Call the function to transform TaskListParams into gen.ListTasksParams
	result, err := toDBListTasksParams(params)

	// Assert: Ensure no error occurred during the transformation
	require.NoError(t, err)

	// Verify that the ListID field was correctly transformed to pgtype.UUID
	require.True(t, result.ID.Valid)
	require.Equal(t, listID[:], result.ID.Bytes[:])

	// Verify that the UserID field was correctly transformed to pgtype.UUID
	require.True(t, result.UserID.Valid)
	require.Equal(t, userID[:], result.UserID.Bytes[:])
}

// TestToDBListOverdueTasksParams tests the toDBListOverdueTasksParams function
func TestToDBListOverdueTasksParams(t *testing.T) {
	// Arrange: Create sample input data
	listID := uuid.New()
	userID := uuid.New()

	// Create the TaskListParams struct
	params := TaskListParams{
		ListID: listID, // Todo List ID
		UserID: userID, // User ID
	}

	// Act: Call the function to transform TaskListParams into gen.ListOverdueTasksParams
	result, err := toDBListOverdueTasksParams(params)

	// Assert: Ensure no error occurred during the transformation
	require.NoError(t, err)

	// Verify that the ListID field was correctly transformed to pgtype.UUID
	require.True(t, result.ListID.Valid)
	require.Equal(t, listID[:], result.ListID.Bytes[:])

	// Verify that the UserID field was correctly transformed to pgtype.UUID
	require.True(t, result.UserID.Valid)
	require.Equal(t, userID[:], result.UserID.Bytes[:])
}

// TestToDBListTasksByStatusParams tests the toDBListTasksByStatusParams function
func TestToDBListTasksByStatusParams(t *testing.T) {
	// Arrange: Create sample input data
	listID := uuid.New()
	userID := uuid.New()
	status := "completed"

	// Create the CountTasksByStatusParams struct
	params := CountTasksByStatusParams{
		ListID: listID,  // Todo List ID
		UserID: userID,  // User ID
		Status: &status, // Status of the tasks
	}

	// Act: Call the function to transform CountTasksByStatusParams into gen.ListTasksByStatusParams
	result, err := toDBListTasksByStatusParams(params)

	// Assert: Ensure no error occurred during the transformation
	require.NoError(t, err)

	// Verify that the ListID field was correctly transformed to pgtype.UUID
	require.True(t, result.ListID.Valid)
	require.Equal(t, listID[:], result.ListID.Bytes[:])

	// Verify that the UserID field was correctly transformed to pgtype.UUID
	require.True(t, result.UserID.Valid)
	require.Equal(t, userID[:], result.UserID.Bytes[:])

	// Verify that the Status field was correctly transformed to pgtype.Text
	require.True(t, result.Status.Valid)
	require.Equal(t, status, result.Status.String)
}

func TestToDBSearchTasksParams(t *testing.T) {
	// Arrange: Create sample input data
	listID := uuid.New()
	userID := uuid.New()
	keyword := "sample search keyword"

	// Create the SearchTasksParams struct
	params := SearchTasksParams{
		ListID:  listID,   // Todo List ID
		UserID:  userID,   // User ID
		Keyword: &keyword, // Search term
	}

	// Act: Call the function to transform SearchTasksParams into gen.SearchTasksParams
	result, err := toDBSearchTasksParams(params)

	// Assert: Ensure no error occurred during the transformation
	require.NoError(t, err)

	// Verify that the ListID field was correctly transformed to pgtype.UUID
	require.True(t, result.ListID.Valid)
	require.Equal(t, listID[:], result.ListID.Bytes[:])

	// Verify that the UserID field was correctly transformed to pgtype.UUID
	require.True(t, result.UserID.Valid)
	require.Equal(t, userID[:], result.UserID.Bytes[:])

	// Verify that the Keyword field was correctly transformed to pgtype.Text
	require.True(t, result.Column3.Valid)
	require.Equal(t, keyword, result.Column3.String)
}

func TestToDBUpdatePriorityParams(t *testing.T) {
	// Arrange: Create sample input data
	taskID := uuid.New()
	listID := uuid.New()
	priority := int32(5)

	// Create the UpdateTaskParams struct
	params := UpdateTaskParams{
		ID:       taskID,    // Task ID
		ListID:   listID,    // List ID
		Priority: &priority, // Priority
	}

	// Act: Call the function to transform UpdateTaskParams into gen.UpdateTaskPriorityParams
	result, err := toDBUpdatePriorityParams(params)

	// Assert: Verify the transformation
	// Verify that TaskID is correctly transformed to pgtype.UUID
	require.NoError(t, err)
	require.True(t, result.ID.Valid)
	require.Equal(t, taskID[:], result.ID.Bytes[:])

	// Verify that ListID is correctly transformed to pgtype.UUID
	require.True(t, result.ListID.Valid)
	require.Equal(t, listID[:], result.ListID.Bytes[:])

	// Verify that Priority is correctly transformed to pgtype.Int4
	require.True(t, result.Priority.Valid)
	require.Equal(t, priority, result.Priority.Int32)
}
