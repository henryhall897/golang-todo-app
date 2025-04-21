package repository

import (
	"fmt"

	"github.com/henryhall897/golang-todo-app/gen/queries/userstore"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
)

func pgToUsers(users userstore.User) (domain.User, error) {
	userID, err := common.FromPgUUID(users.ID)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to transform uuid %w", err)
	}
	userCreatedAt := common.FromPgTimestamp(users.CreatedAt)
	userUpdatedAt := common.FromPgTimestamp(users.UpdatedAt)
	return domain.User{
		ID:        userID,
		Name:      users.Name,
		Email:     users.Email,
		CreatedAt: userCreatedAt,
		UpdatedAt: userUpdatedAt,
	}, nil
}

// toPgListParams converts ListUsersParams to gen.ListUsersParams
func getUsersParamsToPG(params domain.GetUsersParams) userstore.GetUsersParams {
	return userstore.GetUsersParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	}
}

// toPgCreateUserParams converts CreateUserParams to gen.CreateUserParams
func createUserParamsToPG(params domain.CreateUserParams) userstore.CreateUserParams {
	return userstore.CreateUserParams{
		Name:  params.Name,
		Email: params.Email,
	}
}

func updateUserParamsToPG(input domain.UpdateUserParams) (userstore.UpdateUserParams, error) {
	//ToPgUUID converts a UUID to a pgtype.UUID
	pgId, err := common.ToPgUUID(input.ID)
	if err != nil {
		return userstore.UpdateUserParams{}, fmt.Errorf("failed to convert UUID: %w", err)
	}

	userUpdate := userstore.UpdateUserParams{
		ID:    pgId,
		Name:  input.Name,
		Email: input.Email,
	}
	return userUpdate, nil
}
