package tasks

import (
	"time"

	"github.com/google/uuid"
)

// Task represents a task in the database.
type FullTask struct {
	ID          uuid.UUID  `json:"id"`
	ListID      uuid.UUID  `json:"list_id"`
	Title       *string    `json:"title"`
	TaskDesc    *string    `json:"task_desc"`
	Status      *string    `json:"status"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	Priority    *int32     `json:"priority"`
	CompletedAt *time.Time `json:"completed_at"`
}

// CreateTaskParams holds the parameters needed to create a task.
type CreateTaskParams struct {
	ListID   uuid.UUID  `json:"list_id"`
	Title    *string    `json:"title"`
	TaskDesc *string    `json:"description"`
	Status   *string    `json:"status"`
	DueDate  *time.Time `json:"due_date"`
	Priority int32      `json:"priority"`
}

// UpdateTaskParams holds the parameters needed to update a task.
type UpdateTaskParams struct {
	ID          uuid.UUID  `json:"id"`
	ListID      uuid.UUID  `json:"list_id"`
	UserID      uuid.UUID  `json:"user_id"`
	Title       *string    `json:"title"`
	TaskDesc    *string    `json:"task_desc"`
	Status      *string    `json:"status"`
	DueDate     *time.Time `json:"due_date"`
	Priority    *int32     `json:"priority"`
	CompletedAt *time.Time `json:"completed_at"`
}

// MarkTaskCompletedParams holds the parameters needed to mark a task as completed.
type MarkTaskCompletedParams struct {
	ID     uuid.UUID `json:"id"`      // Task ID
	ListID uuid.UUID `json:"list_id"` // Todo List ID
	UserID uuid.UUID `json:"user_id"` // User ID
}

// DeleteTasksParams holds the parameters needed to delete one or more tasks.
type DeleteTasksParams struct {
	IDs    []uuid.UUID `json:"ids"`     // Slice of Task IDs to delete
	ListID uuid.UUID   `json:"list_id"` // Todo List ID
	UserID uuid.UUID   `json:"user_id"` // User ID
}

// TaskListParams holds the parameters needed to list tasks for a specific user and todo list.
type TaskListParams struct {
	ListID uuid.UUID `json:"list_id"` // Todo List ID
	UserID uuid.UUID `json:"user_id"` // User ID
}

// CountTasksByStatusParams holds the parameters needed to count tasks by status for a specific todo list and user.
type CountTasksByStatusParams struct {
	ListID uuid.UUID `json:"list_id"` // Todo List ID
	UserID uuid.UUID `json:"user_id"` // User ID
	Status *string   `json:"status"`  // Status of the tasks (e.g., "completed", "pending", etc.)
}

// SearchTasksParams holds the parameters needed to search tasks for a specific user and todo list.
type SearchTasksParams struct {
	ListID  uuid.UUID `json:"list_id"` // The ID of the todo list
	UserID  uuid.UUID `json:"user_id"` // The ID of the user performing the search
	Keyword *string   `json:"keyword"` // The search term to match in task titles or descriptions
}
