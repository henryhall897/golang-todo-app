// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package gen

import (
	"context"
)

type Querier interface {
	// Bulk delete todo lists for a specific user
	BulkDeleteTodoLists(ctx context.Context, arg BulkDeleteTodoListsParams) (int64, error)
	// Create a new todo list
	CreateTodoList(ctx context.Context, arg CreateTodoListParams) (TodoList, error)
	// Delete a single todo list by ID for a specific user
	DeleteTodoList(ctx context.Context, arg DeleteTodoListParams) (int64, error)
	// Retrieve a todo list by ID, ensuring it belongs to the user
	GetTodoListByID(ctx context.Context, arg GetTodoListByIDParams) (TodoList, error)
	// Retrieve todo lists with pagination
	ListTodoListsWithPagination(ctx context.Context, arg ListTodoListsWithPaginationParams) ([]TodoList, error)
	// Update an existing todo list for a specific user
	UpdateTodoList(ctx context.Context, arg UpdateTodoListParams) (TodoList, error)
}

var _ Querier = (*Queries)(nil)
