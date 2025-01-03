package todolist

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"golang-todo-app/internal/core/common"
	"golang-todo-app/internal/todolist/gen"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		pool: pool,
	}
}

func (s *Store) CreateTodoList(ctx context.Context, userID uuid.UUID, name, description string) (TodoList, error) {
	query := gen.New(s.pool)

	// Use the transform function to convert to database-compatible TodoList
	dbTodoList := toDBTodoListForCreate(userID, name, &description)

	// Prepare the parameters for the query using the transformed struct
	arg := gen.CreateTodoListParams{
		UserID:      dbTodoList.UserID,
		Name:        dbTodoList.Name,
		Description: dbTodoList.Description,
	}

	// Execute the query
	todoList, err := query.CreateTodoList(ctx, arg)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to create todo list: %w", err)
	}

	return toAppTodoList(todoList), nil
}

func (s *Store) GetTodoListByID(ctx context.Context, id, userID uuid.UUID) (TodoList, error) {
	query := gen.New(s.pool)

	// Create the parameter object for the query
	params := gen.GetTodoListByIDParams{
		ID:     id,
		UserID: userID,
	}

	// Execute the query
	todoList, err := query.GetTodoListByID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TodoList{}, common.ErrNotFound
		}
		return TodoList{}, fmt.Errorf("failed to get todo list: %w", err)
	}

	return toAppTodoList(todoList), nil
}

func (s *Store) UpdateTodoList(ctx context.Context, id uuid.UUID, userID uuid.UUID, name, description string) (TodoList, error) {
	query := gen.New(s.pool)

	// Use the transform function to convert to database-compatible TodoList for update
	dbTodoList, err := toDBTodoListForUpdate(id, userID, name, &description)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list for update: %w", err)
	}

	// Prepare the parameters for the query using the transformed struct
	arg := gen.UpdateTodoListParams{
		ID:          dbTodoList.ID,
		UserID:      dbTodoList.UserID,
		Name:        dbTodoList.Name,
		Description: dbTodoList.Description,
	}

	// Execute the query and get the updated record
	updatedTodoList, err := query.UpdateTodoList(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TodoList{}, common.ErrNotFound
		}
		return TodoList{}, fmt.Errorf("failed to update todo list: %w", err)
	}

	return toAppTodoList(updatedTodoList), nil
}

func (s *Store) ListTodoListsWithPagination(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]TodoList, error) {
	query := gen.New(s.pool)

	// Prepare query parameters
	arg := gen.ListTodoListsWithPaginationParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	// Execute the query
	todoLists, err := query.ListTodoListsWithPagination(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []TodoList{}, nil // Return an empty slice if no rows are found
		}
		return nil, fmt.Errorf("failed to list todo lists: %w", err)
	}

	// Transform the results into application-level TodoList structs
	results := make([]TodoList, 0, len(todoLists))
	for _, todo := range todoLists {
		results = append(results, toAppTodoList(todo))
	}

	return results, nil
}

func (s *Store) DeleteTodoList(ctx context.Context, id uuid.UUID, userID uuid.UUID) (int64, error) {
	query := gen.New(s.pool)

	// Create the parameter object for the query
	arg := gen.DeleteTodoListParams{
		ID:     id,
		UserID: userID,
	}

	// Execute the delete query
	rowsAffected, err := query.DeleteTodoList(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, common.ErrNotFound
		}
		return 0, fmt.Errorf("failed to delete todo list: %w", err)
	}

	return rowsAffected, nil
}

func (s *Store) BulkDeleteTodoLists(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (int64, error) {
	query := gen.New(s.pool)

	// Prepare the parameter object for the query
	arg := gen.BulkDeleteTodoListsParams{
		Column1: ids, // Use Column1 instead of IDs
		UserID:  userID,
	}

	// Execute the delete query
	rowsAffected, err := query.BulkDeleteTodoLists(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, common.ErrNotFound
		}
		return 0, fmt.Errorf("failed to bulk delete todo lists: %w", err)
	}

	return rowsAffected, nil
}
