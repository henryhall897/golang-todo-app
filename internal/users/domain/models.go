package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// GetUsersParams defines the parameters for listing users.
type GetUsersParams struct {
	Limit  int
	Offset int
}

type CreateUserParams struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateUserParams represents the parameters for updating a user.
type UpdateUserParams struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}
