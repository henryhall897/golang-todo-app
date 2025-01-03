//go:build unit

package todo_list

import (
	"testing"
	"time"

	"golang-todo-app/internal/todo_list/gen"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestToAppTodoList(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	validUserID := uuid.New()
	validTime := time.Now()

	genTodo := gen.TodoList{
		ID: pgtype.UUID{
			Bytes: validUUID,
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: validUserID,
			Valid: true,
		},
		Name:        "Sample Todo List",
		Description: pgtype.Text{String: "Sample Description", Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: validTime, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: validTime, Valid: true},
	}

	// Act
	result, err := toAppTodoList(genTodo)

	// Assert
	require.NoError(t, err)
	require.Equal(t, validUUID, result.ID)
	require.Equal(t, validUserID, result.UserID)
	require.Equal(t, "Sample Todo List", result.Name)
	require.Equal(t, "Sample Description", result.Description)
	require.Equal(t, validTime, result.CreatedAt)
	require.Equal(t, validTime, result.UpdatedAt)
}

func TestToAppTodoListInvalidID(t *testing.T) {
	// Arrange
	genTodo := gen.TodoList{
		ID: pgtype.UUID{
			Valid: false,
		},
		UserID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		Name:        "Sample Todo List",
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
	genTodo := gen.TodoList{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		UserID: pgtype.UUID{
			Valid: false,
		},
		Name:        "Sample Todo List",
		Description: pgtype.Text{String: "Sample Description", Valid: true},
	}

	// Act
	_, err := toAppTodoList(genTodo)

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid user id")
}

func TestToDBTodoList(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	validUserID := uuid.New()
	description := "Sample Description"

	// Act
	result, err := toDBTodoList(validUUID, validUserID, "Sample Todo List", &description)

	// Assert
	require.NoError(t, err)

	// Convert `pgtype.UUID.Bytes` to `uuid.UUID` for comparison
	resultID, err := uuid.FromBytes(result.ID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validUUID, resultID)

	userID, err := uuid.FromBytes(result.UserID.Bytes[:])
	require.NoError(t, err)
	require.Equal(t, validUserID, userID)

	require.Equal(t, "Sample Todo List", result.Name)
	require.True(t, result.Description.Valid)
	require.Equal(t, "Sample Description", result.Description.String)
}

func TestToDBTodoListInvalidID(t *testing.T) {
	// Arrange
	invalidUUID := uuid.Nil
	validUserID := uuid.New()
	description := "Sample Description"

	// Act
	_, err := toDBTodoListForUpdate(invalidUUID, validUserID, "Sample Todo List", &description)

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid UUID for id")
}

func TestToDBTodoListInvalidUserID(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	invalidUserID := uuid.Nil
	description := "Sample Description"

	// Act
	_, err := toDBTodoList(validUUID, invalidUserID, "Sample Todo List", &description)

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid UUID for user_id")
}

func TestToDBTodoListNilDescription(t *testing.T) {
	// Arrange
	validUUID := uuid.New()
	validUserID := uuid.New()

	// Act
	result, err := toDBTodoList(validUUID, validUserID, "Sample Todo List", nil)

	// Assert
	require.NoError(t, err)
	require.False(t, result.Description.Valid)
}
