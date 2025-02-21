package todolist

import (
	"fmt"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/todolists/gen"

	"github.com/google/uuid"
)

// toAppTodoList transforms a database (SQLC-generated) gen.TodoList to an application-level TodoList.
func toAppTodoList(todo gen.Todolist) (TodoList, error) {
	if !todo.ID.Valid {
		return TodoList{}, fmt.Errorf("invalid todo list id")
	}
	if !todo.UserID.Valid {
		return TodoList{}, fmt.Errorf("invalid user id")
	}

	// Convert pgtype.UUID to uuid.UUID
	id, err := uuid.FromBytes(todo.ID.Bytes[:])
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to parse todo list id: %w", err)
	}

	userID, err := uuid.FromBytes(todo.UserID.Bytes[:])
	if err != nil {
		return TodoList{}, fmt.Errorf("failed to parse user id: %w", err)
	}

	// Transform the TodoList structure
	return TodoList{
		ID:          id,
		UserID:      userID,
		Title:       todo.Title,
		Description: todo.Description.String,
		CreatedAt:   todo.CreatedAt.Time,
		UpdatedAt:   todo.UpdatedAt.Time,
	}, nil
}

// toDBTodoListUpdate transforms UpdateTodoListParams into gen.UpdateTodoListParams
func toDBTodoListUpdate(params UpdateTodoListParams) (gen.UpdateTodoListParams, error) {
	// Convert ID to pgtype.UUID
	dbID, err := common.ToPgUUID(params.ID)
	if err != nil {
		return gen.UpdateTodoListParams{}, fmt.Errorf("failed to convert ID: %w", err)
	}

	// Convert UserID to pgtype.UUID
	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.UpdateTodoListParams{}, fmt.Errorf("failed to convert UserID: %w", err)
	}

	// Convert Description to pgtype.Text
	dbDescription := common.ToPgText(&params.Description)

	return gen.UpdateTodoListParams{
		ID:          dbID,
		UserID:      dbUserID,
		Title:       params.Title,
		Description: dbDescription,
	}, nil
}

// toDBCreateTodoList transforms CreateTodoListParams into gen.CreateTodoListParams
func toDBCreateTodoList(params CreateTodoListParams) (gen.CreateTodoListParams, error) {
	// Convert UserID to pgtype.UUID
	pgUUID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.CreateTodoListParams{}, fmt.Errorf("failed to convert UserID: %w", err)
	}

	// Convert Description to pgtype.Text
	pgDescription := common.ToPgText(&params.Description)

	return gen.CreateTodoListParams{
		UserID:      pgUUID,
		Title:       params.Title,
		Description: pgDescription,
	}, nil
}

// toDBGetTodoListByID transforms GetTodoListByIDParams into gen.GetTodoListByIDParams
func toDBGetTodoListByID(params GetTodoListByIDParams) (gen.GetTodoListByIDParams, error) {
	// Convert ID to pgtype.UUID
	dbID, err := common.ToPgUUID(params.ID)
	if err != nil {
		return gen.GetTodoListByIDParams{}, fmt.Errorf("failed to convert ID: %w", err)
	}

	// Convert UserID to pgtype.UUID
	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.GetTodoListByIDParams{}, fmt.Errorf("failed to convert UserID: %w", err)
	}

	return gen.GetTodoListByIDParams{
		ID:     dbID,
		UserID: dbUserID,
	}, nil
}

// toDBListTodoListsWithPagination transforms ListTodoListsWithPaginationParams into gen.ListTodoListsWithPaginationParams
func toDBListTodoListsWithPagination(params ListTodoListsWithPaginationParams) (gen.ListTodoListsWithPaginationParams, error) {
	// Convert UserID to pgtype.UUID
	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.ListTodoListsWithPaginationParams{}, fmt.Errorf("failed to convert UserID: %w", err)
	}

	return gen.ListTodoListsWithPaginationParams{
		UserID: dbUserID,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// toDBDeleteLists transforms DeleteTodoListsParams into gen.DeleteTodoListsParams using common transforms
func toDBDeleteLists(params DeleteTodoListsParams) (gen.DeleteTodoListsParams, error) {
	// Convert UserID using common transform
	dbUserID, err := common.ToPgUUID(params.UserID)
	if err != nil {
		return gen.DeleteTodoListsParams{}, fmt.Errorf("failed to convert UserID: %w", err)
	}

	// Convert the list of IDs to pgtype.UUIDArray using common transform
	dbIDs, err := common.ToPgUUIDArray(params.IDs)
	if err != nil {
		return gen.DeleteTodoListsParams{}, fmt.Errorf("failed to convert IDs array: %w", err)
	}

	return gen.DeleteTodoListsParams{
		UserID:  dbUserID,
		Column2: dbIDs, // SQLC named the ID array as Column2
	}, nil
}
