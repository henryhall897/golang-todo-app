package handlers

import (
	"context"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users"

	"github.com/google/uuid"
)

// MockStore is a mock implementation of the Store interface.
type MockStore struct {
	CreateUserFunc     func(ctx context.Context, params users.CreateUserParams) (users.User, error)
	GetUserByIDFunc    func(ctx context.Context, id uuid.UUID) (users.User, error)
	GetUserByEmailFunc func(ctx context.Context, email string) (users.User, error)
	ListUsersFunc      func(ctx context.Context, params users.ListUsersParams) ([]users.User, error)
	UpdateUserFunc     func(ctx context.Context, params users.UpdateUserParams) (users.User, error)
	DeleteUserFunc     func(ctx context.Context, id uuid.UUID) error
}

// Mock method implementations for the Store interface.

func (m *MockStore) CreateUser(ctx context.Context, params users.CreateUserParams) (users.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, params)
	}
	return users.User{}, nil
}

func (m *MockStore) GetUserByID(ctx context.Context, id uuid.UUID) (users.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return users.User{}, nil
}

func (m *MockStore) ListUsers(ctx context.Context, params users.ListUsersParams) ([]users.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, params)
	}
	return []users.User{}, nil
}

func (m *MockStore) UpdateUser(ctx context.Context, params users.UpdateUserParams) (users.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, params)
	}
	return users.User{}, common.ErrNotFound
}

func (m *MockStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}
func (m *MockStore) GetUserByEmail(ctx context.Context, email string) (users.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return users.User{}, nil
}

// MockLogger is a mock implementation of the Logger interface.
type MockLogger struct {
	ErrorwFunc func(msg string, keysAndValues ...interface{})
	InfowFunc  func(msg string, keysAndValues ...interface{})
	DebugwFunc func(msg string, keysAndValues ...interface{})
	// Add other methods as needed
}

func (m *MockLogger) Errorw(msg string, keysAndValues ...interface{}) {
	if m.ErrorwFunc != nil {
		m.ErrorwFunc(msg, keysAndValues...)
	}
}

func (m *MockLogger) Infow(msg string, keysAndValues ...interface{}) {
	if m.InfowFunc != nil {
		m.InfowFunc(msg, keysAndValues...)
	}
}

func (m *MockLogger) Debugw(msg string, keysAndValues ...interface{}) {
	if m.DebugwFunc != nil {
		m.DebugwFunc(msg, keysAndValues...)
	}
}
