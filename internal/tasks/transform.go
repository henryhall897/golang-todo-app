package tasks

import (
	"fmt"
	"golang-todo-app/internal/tasks/gen"
	"time"

	"golang-todo-app/internal/core/common"

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
	dbDescription := common.ToPgText(params.TaskDesc)
	dbStatus := common.ToPgText(params.Status)
	dbDueDate := common.ToPgTimestamptz(params.DueDate)
	dbPriority := common.ToPgInt4(params.Priority)

	// Return the transformed Task
	return gen.CreateTaskParams{
		ListID:   dbListID,
		Title:    dbTitle,
		TaskDesc: dbDescription,
		Status:   dbStatus,
		DueDate:  dbDueDate,
		Priority: dbPriority,
	}, nil
}

// toFullTask converts a Task (pgtype-based) struct to a FullTask (Go type-based) struct.
func toFullTask(dbTask gen.Task) (FullTask, error) {
	// Convert ID
	id, err := common.FromPgUUID(dbTask.ID)
	if err != nil {
		return FullTask{}, fmt.Errorf("invalid id: %w", err)
	}

	// Convert ListID
	listID, err := common.FromPgUUID(dbTask.ListID)
	if err != nil {
		return FullTask{}, fmt.Errorf("invalid list_id: %w", err)
	}

	// Convert Description (check for nil)
	var description *string
	if dbTask.TaskDesc.Valid {
		description = &dbTask.TaskDesc.String
	}

	title := common.FromPgText(dbTask.Title)

	// Convert Status (check for nil)
	var status *string
	if dbTask.Status.Valid {
		status = &dbTask.Status.String
	}

	// Convert DueDate (check for nil)
	var dueDate *time.Time
	if dbTask.DueDate.Valid {
		dueDate = &dbTask.DueDate.Time
	}

	// Convert CreatedAt and UpdatedAt
	createdAt := dbTask.CreatedAt.Time
	updatedAt := dbTask.UpdatedAt.Time

	// Convert CompletedAt (check for nil)
	var completedAt *time.Time
	if dbTask.CompletedAt.Valid {
		completedAt = &dbTask.CompletedAt.Time
	}

	// Convert Priority (check if valid)
	var priority *int32
	if dbTask.Priority.Valid {
		priority = common.Ptr(dbTask.Priority.Int32)
	}

	// Return the transformed FullTask
	return FullTask{
		ID:          id,
		ListID:      listID,
		Title:       title,
		TaskDesc:    description,
		Status:      status,
		DueDate:     dueDate,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
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
	// Convert and validate TaskID
	dbTaskID, err := common.ToPgUUID(params.ID)
	if err != nil {
		return gen.UpdateTaskParams{}, fmt.Errorf("invalid task_id: %w", err)
	}

	// Convert and validate UserID
	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.UpdateTaskParams{}, fmt.Errorf("invalid user_id: %w", err)
	}

	// Handle Title
	var dbTitle pgtype.Text
	if params.Title != nil {
		dbTitle = common.ToPgText(params.Title)
	} else {
		dbTitle.Valid = false
	}

	// Convert Task Description
	var dbTaskDesc pgtype.Text
	if params.TaskDesc != nil {
		dbTaskDesc = common.ToPgText(params.TaskDesc)
	} else {
		dbTaskDesc.Valid = false
	}

	// Convert Status
	var dbStatus pgtype.Text
	if params.Status != nil {
		dbStatus = common.ToPgText(params.Status)
	} else {
		dbStatus.Valid = false
	}

	// Convert DueDate
	var dbDueDate pgtype.Timestamptz
	if params.DueDate != nil {
		dbDueDate = common.ToPgTimestamptz(params.DueDate)
	} else {
		dbDueDate.Valid = false
	}

	// Convert CompletedAt
	var dbCompletedAt pgtype.Timestamptz
	if params.CompletedAt != nil {
		dbCompletedAt = common.ToPgTimestamptz(params.CompletedAt)
	} else {
		dbCompletedAt.Valid = false
	}

	// Handle Priority (if it's not nil, we need to handle it properly)
	var dbPriority pgtype.Int4

	if params.CompletedAt != nil {
		// If the task is marked as completed, explicitly set priority to NULL
		dbPriority.Valid = false
	} else if params.Priority != nil {
		// If priority is provided, use it
		dbPriority = common.ToPgInt4(*params.Priority)
	} else {
		// If priority is not provided and task is not completed, leave it unchanged
		dbPriority.Valid = false // Mark as invalid to retain the current value
	}

	// Return the transformed struct
	return gen.UpdateTaskParams{
		ID:          dbTaskID,
		UserID:      dbUserID,
		Title:       dbTitle,
		TaskDesc:    dbTaskDesc,
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

func toDBUpdatePriorityParams(params UpdateTaskParams) gen.UpdateTaskPriorityParams {
	// Convert and validate TaskID
	dbTaskID, _ := common.ToPgUUID(params.ID)

	// Convert and validate UserID
	dbListID, _ := common.ToPgUUID(params.ListID)

	// Convert Priority
	dbPriority := common.ToPgInt4(*params.Priority)

	// Return the transformed struct
	return gen.UpdateTaskPriorityParams{
		ID:       dbTaskID,
		ListID:   dbListID,
		Priority: dbPriority,
	}
}
