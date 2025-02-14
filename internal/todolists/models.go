package todolist

import (
	"time"

	"github.com/google/uuid"
)

type TodoList struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTodoListParams struct {
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

// GetTodoListByIDParams holds the parameters for fetching a todo list by ID
type GetTodoListByIDParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

// UpdateTodoListParams holds the parameters for updating a todo list
type UpdateTodoListParams struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

// ListTodoListsWithPaginationParams holds the parameters for paginated todo list retrieval
type ListTodoListsWithPaginationParams struct {
	UserID uuid.UUID `json:"user_id"`
	Limit  int32     `json:"limit"`
	Offset int32     `json:"offset"`
}

// DeleteTodoListsParams holds the parameters for deleting todo lists
type DeleteTodoListsParams struct {
	UserID uuid.UUID   `json:"user_id"`
	IDs    []uuid.UUID `json:"ids"` // List of IDs to delete (nil for deleting all)
}
