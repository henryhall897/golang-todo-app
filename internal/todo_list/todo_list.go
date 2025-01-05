package todolist

import (
	"context"
	"errors"
	"fmt"

	"golang-todo-app/internal/core/common"
	"golang-todo-app/internal/todo_list/gen"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

func (s *Store) CreateTodoList(ctx context.Context, userID uuid.UUID, name, TodoDesc string) (TodoList, error) {
	query := gen.New(s.pool)

	// Use the transform function to convert to database-compatible TodoList
	dbTodoList, err := toDBTodoListForCreate(userID, name, &TodoDesc)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list: %w", err)
	}

	// Prepare the parameters for the query using the transformed struct
	arg := gen.CreateTodoListParams{
		UserID:   dbTodoList.UserID,
		Name:     dbTodoList.Name,
		TodoDesc: dbTodoList.TodoDesc,
	}

	// Execute the query
	todoList, err := query.CreateTodoList(ctx, arg)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to create todo list: %w", err)
	}

	result, err := toAppTodoList(todoList)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list: %w", err)
	}

	return result, nil
}

func (s *Store) GetTodoListByID(ctx context.Context, id, userID uuid.UUID) (TodoList, error) {
	query := gen.New(s.pool)

	// Convert UUIDs using the transform function
	dbID, err := toDBUUID(id)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list id: %w", err)
	}

	dbUserID, err := toDBUUID(userID)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform user id: %w", err)
	}

	// Create the parameter object for the query
	params := gen.GetTodoListByIDParams{
		ID:     dbID,
		UserID: dbUserID,
	}

	// Execute the query
	todoList, err := query.GetTodoListByID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TodoList{}, common.ErrNotFound
		}
		return TodoList{}, fmt.Errorf("failed to get todo list: %w", err)
	}

	// Transform to application-level model
	result, err := toAppTodoList(todoList)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list: %w", err)
	}

	return result, nil
}

func (s *Store) UpdateTodoList(ctx context.Context, id uuid.UUID, userID uuid.UUID, name, TodoDesc string) (TodoList, error) {
	query := gen.New(s.pool)

	// Use the transform function to convert to database-compatible TodoList for update
	dbTodoList, err := toDBTodoListForUpdate(id, userID, name, &TodoDesc)
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to transform todo list for update: %w", err)
	}

	// Prepare the parameters for the query using the transformed struct
	arg := gen.UpdateTodoListParams{
		ID:       dbTodoList.ID,
		UserID:   dbTodoList.UserID,
		Name:     dbTodoList.Name,
		TodoDesc: dbTodoList.TodoDesc,
	}

	// Execute the query and get the updated record
	updatedTodoList, err := query.UpdateTodoList(ctx, arg)
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

func (s *Store) ListTodoListsWithPagination(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]TodoList, error) {
	query := gen.New(s.pool)

	// Convert userID to database-compatible UUID
	dbUserID, err := toDBUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to transform user ID: %w", err)
	}

	// Prepare query parameters
	arg := gen.ListTodoListsWithPaginationParams{
		UserID: dbUserID,
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
		result, err := toAppTodoList(todo)
		if err != nil {
			return nil, fmt.Errorf("failed to transform todo list: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (s *Store) DeleteTodoList(ctx context.Context, id uuid.UUID, userID uuid.UUID) (int64, error) {
	query := gen.New(s.pool)

	// Convert UUIDs using the transform function
	dbID, err := toDBUUID(id)
	if err != nil {
		return 0, fmt.Errorf("failed to transform todo list id: %w", err)
	}

	dbUserID, err := toDBUUID(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to transform user id: %w", err)
	}

	// Create the parameter object for the query
	arg := gen.DeleteTodoListParams{
		ID:     dbID,
		UserID: dbUserID,
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

func (s *Store) BulkDeleteTodoLists(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error) {
	query := gen.New(s.pool)

	// Convert userID to database-compatible UUID
	dbUserID, err := toDBUUID(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to transform user id: %w", err)
	}

	// Convert ids slice to database-compatible UUIDs
	dbIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		dbID, err := toDBUUID(id)
		if err != nil {
			return 0, fmt.Errorf("failed to transform todo list id: %w", err)
		}
		dbIDs[i] = dbID
	}

	// Prepare the parameter object for the query
	arg := gen.BulkDeleteTodoListsParams{
		Column1: dbIDs, // Use Column1 instead of IDs
		UserID:  dbUserID,
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
