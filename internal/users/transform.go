package users

import (
	"fmt"
	"golang-todo-app/internal/users/gen"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

// uuidToPgUUID converts a uuid.UUID into a pgtype.UUID
func uuidToPgUUID(id uuid.UUID) (pgtype.UUID, error) {
	if id == uuid.Nil {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID: UUID is nil")
	}

	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}, nil
}

// toPgListParams converts ListUsersParams to gen.ListUsersParams
func toPgListParams(params ListUsersParams) gen.ListUsersParams {
	return gen.ListUsersParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	}
}
