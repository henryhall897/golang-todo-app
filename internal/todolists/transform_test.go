//go:build unit

package todolist

import (
	"testing"
	"time"

	"github.com/henryhall897/golang-todo-app/internal/todolists/gen"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestToAppTodoList(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	validUserID := uuid.New()
	validTime := time.Now()

	genTodo := gen.Todolist{
		ID: pgtype.UUID{
			Bytes: validUUID,
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: validUserID,
			Valid: true,
		},
		Title:       "Sample Todo List",
		Description: pgtype.Text{String: "Sample Description", Valid: true},
		// Mock pgtype.Timestamptz values to simulate database output.
		CreatedAt: pgtype.Timestamp{Time: validTime, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: validTime, Valid: true},
	}

	// Act
	result, err := toAppTodoList(genTodo)

	// Assert
	require.NoError(t, err)
	require.Equal(t, validUUID, result.ID)
	require.Equal(t, validUserID, result.UserID)
	require.Equal(t, "Sample Todo List", result.Title)
	require.Equal(t, "Sample Description", result.Description)
	require.Equal(t, validTime, result.CreatedAt)
	require.Equal(t, validTime, result.UpdatedAt)
}

func TestToAppTodoListInvalidID(t *testing.T) {
	// Arrange
	genTodo := gen.Todolist{
		ID: pgtype.UUID{
			Valid: false,
		},
		UserID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		Title:       "Sample Todo List",
		Description: pgtype.Text{String: "Sample Description", Valid: true},
	}

	// Act
	_, err := toAppTodoList(genTodo)

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid todo list id")
}

func TestToAppTodoListInvalidUserID(t *testing.T) {
	// Arrange
	genTodo := gen.Todolist{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		UserID: pgtype.UUID{
			Valid: false,
		},
		Title:       "Sample Todo List",
		Description: pgtype.Text{String: "Sample Description", Valid: true},
	}

	// Act
	_, err := toAppTodoList(genTodo)

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid user id")
}

func TestToDBTodoListInvalidID(t *testing.T) {
	// Arrange
	invalidUUID := uuid.Nil
	validUserID := uuid.New()
	description := "Sample Description"

	params := UpdateTodoListParams{
		ID:          invalidUUID,
		UserID:      validUserID,
		Title:       "Sample Todo List",
		Description: description,
	}

	// Act
	_, err := toDBTodoListUpdate(params)

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to convert ID: invalid UUID: cannot be nil")
}

func TestToDBCreateTodoList(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	description := "Sample Description"

	params := CreateTodoListParams{
		UserID:      validUUID,
		Title:       "Sample Todo List",
		Description: description,
	}

	// Act
	result, err := toDBCreateTodoList(params)

	// Assert
	require.NoError(t, err)

	// Convert `pgtype.UUID.Bytes` to `uuid.UUID` for comparison
	resultUserID, err := uuid.FromBytes(result.UserID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validUUID, resultUserID)

	require.Equal(t, "Sample Todo List", result.Title)
	require.True(t, result.Description.Valid)
	require.Equal(t, "Sample Description", result.Description.String)
}

func TestToDBGetTodoListByID(t *testing.T) {
	// Arrange
	validID := uuid.New()
	validUserID := uuid.New()

	params := GetTodoListByIDParams{
		ID:     validID,
		UserID: validUserID,
	}

	// Act
	result, err := toDBGetTodoListByID(params)

	// Assert
	require.NoError(t, err)

	// Convert `pgtype.UUID.Bytes` to `uuid.UUID` for comparison
	resultID, err := uuid.FromBytes(result.ID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validID, resultID)

	resultUserID, err := uuid.FromBytes(result.UserID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validUserID, resultUserID)
}

func TestToDBTodoListUpdate_Success(t *testing.T) {
	// Arrange
	validID := uuid.New()
	validUserID := uuid.New()
	description := "Updated Description"

	params := UpdateTodoListParams{
		ID:          validID,
		UserID:      validUserID,
		Title:       "Updated Todo List",
		Description: description,
	}

	// Act
	result, err := toDBTodoListUpdate(params)

	// Assert
	require.NoError(t, err)

	// Convert `pgtype.UUID.Bytes` to `uuid.UUID` for comparison
	resultID, err := uuid.FromBytes(result.ID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validID, resultID)

	resultUserID, err := uuid.FromBytes(result.UserID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validUserID, resultUserID)

	require.Equal(t, "Updated Todo List", result.Title)
	require.True(t, result.Description.Valid)
	require.Equal(t, "Updated Description", result.Description.String)
}

func TestToDBListTodoListsWithPagination_Success(t *testing.T) {
	// Arrange
	validUserID := uuid.New()
	limit := int32(10)
	offset := int32(5)

	params := ListTodoListsWithPaginationParams{
		UserID: validUserID,
		Limit:  limit,
		Offset: offset,
	}

	// Act
	result, err := toDBListTodoListsWithPagination(params)

	// Assert
	require.NoError(t, err)

	// Convert `pgtype.UUID.Bytes` to `uuid.UUID` for comparison
	resultUserID, err := uuid.FromBytes(result.UserID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validUserID, resultUserID)

	// Ensure Limit and Offset are correctly assigned
	require.Equal(t, limit, result.Limit)
	require.Equal(t, offset, result.Offset)
}

func TestToDBDeleteLists_Success(t *testing.T) {
	// Arrange
	validUserID := uuid.New()
	validIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	params := DeleteTodoListsParams{
		UserID: validUserID,
		IDs:    validIDs,
	}

	// Act
	result, err := toDBDeleteLists(params)

	// Assert
	require.NoError(t, err)

	// Convert `pgtype.UUID.Bytes` to `uuid.UUID` for comparison
	resultUserID, err := uuid.FromBytes(result.UserID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validUserID, resultUserID)

	// Ensure IDs are correctly assigned and converted
	require.Len(t, result.Column2, len(validIDs))
	for i, pgUUID := range result.Column2 {
		require.True(t, pgUUID.Valid)
		require.Equal(t, validIDs[i][:], pgUUID.Bytes[:])
	}
}

func TestToDBDeleteLists_DeleteAll(t *testing.T) {
	// Arrange
	validUserID := uuid.New()

	params := DeleteTodoListsParams{
		UserID: validUserID,
		IDs:    nil, // This should trigger "delete all"
	}

	// Act
	result, err := toDBDeleteLists(params)

	// Assert
	require.NoError(t, err)

	// Convert `pgtype.UUID.Bytes` to `uuid.UUID` for comparison
	resultUserID, err := uuid.FromBytes(result.UserID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validUserID, resultUserID)

	// Ensure IDs array is nil, meaning all should be deleted
	require.Nil(t, result.Column2)
}
