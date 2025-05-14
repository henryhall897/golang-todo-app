package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetUsersParams defines the parameters for listing users.
type GetUsersParams struct {
	Limit  int
	Offset int
}

type CreateUserParams struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// UpdateUserParams represents the parameters for updating a user.
type UpdateUserParams struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type UpdateUserRoleParams struct {
	ID   uuid.UUID `json:"id"`
	Role string    `json:"role"`
}

// AllRoles is a registry of valid roles
var AllRoles = map[Role]struct{}{
	RoleUser:  {},
	RoleAdmin: {},
}
