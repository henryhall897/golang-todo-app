package users

import (
	"fmt"
	"golang-todo-app/internal/users/gen"

	"github.com/google/uuid"
)

func pgToUsers(users gen.User) (User, error) {
	if !users.ID.Valid {
		return User{}, fmt.Errorf("invalid user id")
	}
	id, err := uuid.Parse(users.ID.String())
	if err != nil {
		return User{}, fmt.Errorf("failed to parse uuid")
	}

	return User{
		ID:        id,
		Name:      users.Name,
		Email:     users.Email,
		CreatedAt: users.CreatedAt.Time,
		UpdatedAt: users.UpdatedAt.Time,
	}, nil
}
