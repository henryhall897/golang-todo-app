package todolist

import (
	"context"
	"errors"
	"fmt"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/todolists/gen"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		pool: pool,
	}
}

func (s *Store) CreateTodoList(ctx context.Context, params CreateTodoListParams) (TodoList, error) {
	query := gen.New(s.pool)

	// Transform params to database-compatible struct
	dbTodoList, err := toDBCreateTodoList(params)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list: %w", err)
	}

	// Execute the query
	todoList, err := query.CreateTodoList(ctx, dbTodoList)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to create todo list: %w", err)
	}

	// Transform database model to application model
	result, err := toAppTodoList(todoList)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list: %w", err)
	}

	return result, nil
}

func (s *Store) GetTodoListByID(ctx context.Context, params GetTodoListByIDParams) (TodoList, error) {
	query := gen.New(s.pool)

	// Transform params to database-compatible struct
	dbParams, err := toDBGetTodoListByID(params)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list params: %w", err)
	}

	// Execute the query
	todoList, err := query.GetTodoListByID(ctx, dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TodoList{}, common.ErrNotFound
		}
		return TodoList{}, fmt.Errorf("failed to get todo list: %w", err)
	}

	// Transform database model to application model
	result, err := toAppTodoList(todoList)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list: %w", err)
	}

	return result, nil
}

func (s *Store) UpdateTodoList(ctx context.Context, params UpdateTodoListParams) (TodoList, error) {
	query := gen.New(s.pool)

	// Transform params to database-compatible struct
	dbParams, err := toDBTodoListUpdate(params)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list for update: %w", err)
	}

	// Execute the query and get the updated record
	updatedTodoList, err := query.UpdateTodoList(ctx, dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TodoList{}, common.ErrNotFound
		}
		return TodoList{}, fmt.Errorf("failed to update todo list: %w", err)
	}

	// Transform the updated record to the application-level model
	result, err := toAppTodoList(updatedTodoList)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform updated todo list: %w", err)
	}

	return result, nil
}

func (s *Store) ListTodoListsWithPagination(ctx context.Context, params ListTodoListsWithPaginationParams) ([]TodoList, error) {
	query := gen.New(s.pool)

	// Transform params to database-compatible struct
	dbParams, err := toDBListTodoListsWithPagination(params)
	if err != nil {
		return nil, fmt.Errorf("failed to transform pagination params: %w", err)
	}

	// Execute the query
	todoLists, err := query.ListTodoListsWithPagination(ctx, dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []TodoList{}, nil // Return an empty slice if no rows are found
		}
		return nil, fmt.Errorf("failed to list todo lists: %w", err)
	}

	// Transform the results into application-level TodoList structs
	results := make([]TodoList, 0, len(todoLists))
	for _, todo := range todoLists {
		result, err := toAppTodoList(todo)
		if err != nil {
			return nil, fmt.Errorf("failed to transform todo list: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (s *Store) DeleteTodoLists(ctx context.Context, params DeleteTodoListsParams) (int64, error) {
	query := gen.New(s.pool)

	// Transform params to database-compatible struct
	dbParams, err := toDBDeleteLists(params)
	if err != nil {
		return 0, fmt.Errorf("failed to transform delete todo lists params: %w", err)
	}

	// Execute the delete query
	rowsAffected, err := query.DeleteTodoLists(ctx, dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, common.ErrNotFound
		}
		return 0, fmt.Errorf("failed to delete todo lists: %w", err)
	}

	return rowsAffected, nil
}
