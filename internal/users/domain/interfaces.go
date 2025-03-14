package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the methods required for user operations.
//
//go:generate moq -out=../../../gen/mocks/usersmock/user_repo_mock.go -pkg=usersmock . Repository
type Repository interface {
	CreateUser(ctx context.Context, newUserParams CreateUserParams) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUsers(ctx context.Context, params GetUsersParams) ([]User, error)
	UpdateUser(ctx context.Context, updateUserparams UpdateUserParams) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

//go:generate moq -out=../../../gen/mocks/usersmock/user_service_mock.go -pkg=usersmock . Service
type Service interface {
	CreateUser(ctx context.Context, params CreateUserParams) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUsers(ctx context.Context, params GetUsersParams) ([]User, error)
	UpdateUser(ctx context.Context, params UpdateUserParams) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
