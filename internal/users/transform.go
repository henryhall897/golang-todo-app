package users

import (
	"fmt"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users/gen"
)

func dbToUsers(users gen.User) (User, error) {
	if !users.ID.Valid {
		return User{}, fmt.Errorf("invalid user id")
	}
	userID, err := common.FromPgUUID(users.ID)
	if err != nil {
		return User{}, fmt.Errorf("failed to parse uuid")
	}
	userCreatedAt := common.FromPgTimestamp(users.CreatedAt)
	userUpdatedAt := common.FromPgTimestamp(users.UpdatedAt)

	return User{
		ID:        userID,
		Name:      users.Name,
		Email:     users.Email,
		CreatedAt: userCreatedAt,
		UpdatedAt: userUpdatedAt,
	}, nil
}

// toPgListParams converts ListUsersParams to gen.ListUsersParams
func toDBListParams(params ListUsersParams) gen.ListUsersParams {
	return gen.ListUsersParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	}
}

// toPgCreateUserParams converts CreateUserParams to gen.CreateUserParams
func toDBCreateUserParams(params CreateUserParams) gen.CreateUserParams {
	return gen.CreateUserParams{
		Name:  params.Name,
		Email: params.Email,
	}
}

func toDBUpdateUserParams(input UpdateUserParams) (gen.UpdateUserParams, error) {
	pgId, err := common.ToPgUUID(input.ID)
	if err != nil {
		return gen.UpdateUserParams{}, fmt.Errorf("failed to convert UUID: %w", err)
	}

	userUpdate := gen.UpdateUserParams{
		ID:    pgId,
		Name:  input.Name,
		Email: input.Email,
	}
	return userUpdate, nil
}
