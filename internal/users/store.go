// In internal/users/store.go
package users

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the methods required for user operations.
type Store interface {
	CreateUser(ctx context.Context, name, email string) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	ListUsers(ctx context.Context, params ListUsersParams) ([]User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, name, email string) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
