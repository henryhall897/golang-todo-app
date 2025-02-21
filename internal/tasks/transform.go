package tasks

import (
	"fmt"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/tasks/gen"

	"github.com/jackc/pgx/v5/pgtype"
)

// toDBTask converts CreateTaskParams (Go struct) into a pgtype-compatible Task struct.
func toDBCreateTask(params CreateTaskParams) (gen.CreateTaskParams, error) {
	// Convert and validate ListID
	dbListID, err := common.ToPgUUID(params.ListID)
	if err != nil {
		return gen.CreateTaskParams{}, fmt.Errorf("invalid list_id: %w", err)
	}

	// Convert other fields using common utility functions
	dbTitle := common.ToPgText(params.Title)
	dbDescription := common.ToPgText(params.Description)
	dbStatus := common.ToPgText(params.Status)
	dbDueDate := common.ToPgTimestamp(params.DueDate)
	dbPriority := common.ToPgInt4(params.Priority)

	// Return the transformed Task
	return gen.CreateTaskParams{
		ListID:      dbListID,
		Title:       dbTitle,
		Description: dbDescription,
		Status:      dbStatus,
		DueDate:     dbDueDate,
		Priority:    dbPriority,
	}, nil
}

// toFullTask converts a Task (pgtype-based) struct to a FullTask (Go type-based) struct.
func toFullTask(dbTask gen.Task) (FullTask, error) {
	// Convert ID and ListID using common utility functions
	id, err := common.FromPgUUID(dbTask.ID)
	if err != nil {
		return FullTask{}, fmt.Errorf("invalid id: %w", err)
	}

	listID, err := common.FromPgUUID(dbTask.ListID)
	if err != nil {
		return FullTask{}, fmt.Errorf("invalid list_id: %w", err)
	}

	// Convert fields using utility functions
	title := common.FromPgText(dbTask.Title)
	description := common.FromPgText(dbTask.Description)
	status := common.FromPgText(dbTask.Status)
	dueDate := common.FromPgTimestamp(dbTask.DueDate)
	createdAt := dbTask.CreatedAt.Time
	updatedAt := dbTask.UpdatedAt.Time
	completedAt := common.FromPgTimestamp(dbTask.CompletedAt)
	priority := common.FromPgInt4(dbTask.Priority)

	// Return the transformed FullTask
	return FullTask{
		ID:          id,
		ListID:      listID,
		Title:       title,
		Description: description,
		Status:      status,
		DueDate:     dueDate,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Priority:    priority,
		CompletedAt: completedAt,
	}, nil
}

// toFullTaskList converts a slice of Tasks (pgtype-based) into a slice of FullTasks (Go type-based).
func toFullTaskList(dbTasks []gen.Task) ([]FullTask, error) {
	var tasks []FullTask
	for _, dbTask := range dbTasks {
		task, err := toFullTask(dbTask)
		if err != nil {
			return nil, fmt.Errorf("failed to transform task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// toDBUpdateTaskParams converts UpdateTaskParams (Go struct) into a pgtype-compatible UpdateTaskParams struct.
func toDBUpdateTaskParams(params UpdateTaskParams) (gen.UpdateTaskParams, error) {
	// Convert ID and UserID using utility functions
	dbTaskID, err := common.ToPgUUID(params.ID)
	if err != nil {
		return gen.UpdateTaskParams{}, fmt.Errorf("invalid task_id: %w", err)
	}

	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.UpdateTaskParams{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Use common conversion functions for all other fields
	dbTitle := common.ToPgText(params.Title)
	dbTaskDesc := common.ToPgText(params.Description)
	dbStatus := common.ToPgText(params.Status)
	dbDueDate := common.ToPgTimestamp(params.DueDate)
	dbCompletedAt := common.ToPgTimestamp(params.CompletedAt)

	// Handle Priority
	var dbPriority pgtype.Int4
	if params.CompletedAt != nil {
		// If the task is marked as completed, explicitly set priority to NULL
		dbPriority.Valid = false
	} else if params.Priority != nil {
		// If priority is provided, use the conversion function
		dbPriority = common.ToPgInt4(*params.Priority)
	} else {
		// If priority is not provided and task is not completed, leave it unchanged
		dbPriority.Valid = false
	}

	// Return the transformed struct
	return gen.UpdateTaskParams{
		ID:          dbTaskID,
		UserID:      dbUserID,
		Title:       dbTitle,
		Description: dbTaskDesc,
		Status:      dbStatus,
		DueDate:     dbDueDate,
		Priority:    dbPriority,
		CompletedAt: dbCompletedAt,
	}, nil
}

func toMarkTaskCompletedParams(params gen.UpdateTaskParams) (gen.MarkTaskCompletedParams, error) {

	// Return the transformed MarkTaskCompletedParams
	return gen.MarkTaskCompletedParams{
		ID:     params.ID,
		UserID: params.UserID,
	}, nil
}

// toDBDeleteTasksParams converts DeleteTasksParams (Go struct) into a pgtype-compatible struct.
func toDBDeleteTasksParams(params DeleteTasksParams) (gen.DeleteTasksParams, error) {
	// Convert and validate UserID
	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.DeleteTasksParams{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Convert Task IDs slice
	var dbTaskIDs []pgtype.UUID
	for _, id := range params.IDs {
		dbID, err := common.ToPgUUID(id)
		if err != nil {
			return gen.DeleteTasksParams{}, fmt.Errorf("invalid task_id: %w", err)
		}
		dbTaskIDs = append(dbTaskIDs, dbID)
	}

	// Return the transformed struct
	return gen.DeleteTasksParams{
		Column1: dbTaskIDs,
		UserID:  dbUserID,
	}, nil
}

// toDBListTasksParams converts ListTasksParams (Go struct) into a pgtype-compatible ListTasksParams struct.
func toDBListTasksParams(params TaskListParams) (gen.ListTasksParams, error) {
	// Convert and validate ListID
	dbListID, err := common.ToPgUUID(params.ListID)
	if err != nil {
		return gen.ListTasksParams{}, fmt.Errorf("invalid list_id: %w", err)
	}

	// Convert and validate UserID
	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.ListTasksParams{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Return the transformed struct
	return gen.ListTasksParams{
		ID:     dbListID,
		UserID: dbUserID,
	}, nil
}

// toDBListOverdueTasksParams converts TaskListParams (Go struct) into a pgtype-compatible ListOverdueTasksParams struct.
func toDBListOverdueTasksParams(params TaskListParams) (gen.ListOverdueTasksParams, error) {
	// Convert and validate ListID
	dbListID, err := common.ToPgUUID(params.ListID)
	if err != nil {
		return gen.ListOverdueTasksParams{}, fmt.Errorf("invalid list_id: %w", err)
	}

	// Convert and validate UserID
	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.ListOverdueTasksParams{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Return the transformed struct
	return gen.ListOverdueTasksParams{
		ListID: dbListID,
		UserID: dbUserID,
	}, nil
}

// toDBListTasksByStatusParams converts CountTasksByStatusParams (Go struct) into a pgtype-compatible struct for ListTasksByStatus.
func toDBListTasksByStatusParams(params CountTasksByStatusParams) (gen.ListTasksByStatusParams, error) {
	// Convert and validate ListID
	dbListID, err := common.ToPgUUID(params.ListID)
	if err != nil {
		return gen.ListTasksByStatusParams{}, fmt.Errorf("invalid list_id: %w", err)
	}

	// Convert and validate UserID
	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.ListTasksByStatusParams{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Convert Status to pgtype.Text
	dbStatus := common.ToPgText(params.Status)

	// Return the transformed struct for ListTasksByStatus query
	return gen.ListTasksByStatusParams{
		ListID: dbListID,
		UserID: dbUserID,
		Status: dbStatus,
	}, nil
}

// toDBSearchTasksParams transforms SearchTasksParams into a database-compatible struct.
func toDBSearchTasksParams(params SearchTasksParams) (gen.SearchTasksParams, error) {
	// Convert ListID and UserID
	dbListID, err := common.ToPgUUID(params.ListID)
	if err != nil {
		return gen.SearchTasksParams{}, fmt.Errorf("invalid list_id: %w", err)
	}

	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.SearchTasksParams{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Use common helper to safely convert Keyword to Text
	dbKeyword := common.ToPgText(params.Keyword)

	// Return the transformed struct
	return gen.SearchTasksParams{
		ListID:  dbListID,
		UserID:  dbUserID,
		Column3: dbKeyword,
	}, nil
}

func toDBUpdatePriorityParams(params UpdateTaskParams) (gen.UpdateTaskPriorityParams, error) {
	// Convert and validate TaskID
	dbTaskID, err := common.ToPgUUID(params.ID)
	if err != nil {
		return gen.UpdateTaskPriorityParams{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Convert and validate UserID
	dbListID, err := common.ToPgUUID(params.ListID)
	if err != nil {
		return gen.UpdateTaskPriorityParams{}, fmt.Errorf("invalid list_id: %w", err)
	}

	// Convert Priority
	dbPriority := common.ToPgInt4(*params.Priority)

	// Return the transformed struct
	return gen.UpdateTaskPriorityParams{
		ID:       dbTaskID,
		ListID:   dbListID,
		Priority: dbPriority,
	}, nil
}
