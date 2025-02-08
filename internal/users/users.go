package users

import (
	"context"
	"errors"
	"fmt"

	"golang-todo-app/internal/core/common"
	"golang-todo-app/internal/users/gen"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *UserStore {
	return &UserStore{
		pool: pool,
	}
}

func (s *UserStore) CreateUser(ctx context.Context, name, email string) (User, error) {
	query := gen.New(s.pool)

	user, err := query.CreateUser(ctx, gen.CreateUserParams{
		Name:  name,
		Email: email,
	})
	if err != nil {
		return User{}, err
	}

	result, err := pgToUsers(user)
	if err != nil {
		return User{}, err
	}

	return result, nil
}

func (s *UserStore) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	query := gen.New(s.pool)

	user, err := query.GetUserByID(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, common.ErrNotFound
	} else if err != nil {
		return User{}, &common.UserIDNotFoundError{UserID: id}
	}

	result, err := pgToUsers(user)
	if err != nil {
		return User{}, err
	}

	return result, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (User, error) {
	query := gen.New(s.pool)

	user, err := query.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, common.ErrNotFound
	} else if err != nil {
		return User{}, fmt.Errorf("failed to get user by email %w", err)
	}

	result, err := pgToUsers(user)
	if err != nil {
		return User{}, err
	}

	return result, nil
}

func (s *UserStore) ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error) {
	query := gen.New(s.pool)
	// Convert the ListUsersParams to gen.ListUsersParams
	pgArg := toPgListParams(arg)

	// Execute the query to get users
	users, err := query.ListUsers(ctx, pgArg)
	if errors.Is(err, pgx.ErrNoRows) {
		// Return an empty array if no rows are found
		return []User{}, nil
	} else if err != nil {
		return []User{}, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert the raw database results into the User type
	results := make([]User, 0, len(users))
	for _, u := range users {
		result, err := pgToUsers(u)
		if err != nil {
			return []User{}, fmt.Errorf("failed to convert user: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (s *UserStore) UpdateUser(ctx context.Context, id uuid.UUID, name, email string) (User, error) {
	query := gen.New(s.pool)
	pgId, err := uuidToPgUUID(id)
	if err != nil {
		return User{}, fmt.Errorf("failed to covert uuid to pguuid")
	}

	// Create the parameters for the query
	arg := gen.UpdateUserParams{
		ID:    pgId,
		Name:  name,
		Email: email,
	}

	// Execute the update query
	err = query.UpdateUser(ctx, arg)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, common.ErrNotFound
	} else if err != nil {
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}

	// Retrieve the updated user to return it
	updatedUser, err := s.GetUserByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("failed to retrieve updated user: %w", err)
	}

	return updatedUser, nil
}

func (s *UserStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := gen.New(s.pool)

	// Convert uuid.UUID to pgtype.UUID
	pgId, err := uuidToPgUUID(id)
	if err != nil {
		return fmt.Errorf("failed to convert UUID to pgtype.UUID: %w", err)
	}

	// Execute the delete query
	rowsAffected, err := query.DeleteUser(ctx, pgId)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}

	if rowsAffected == 0 {
		return common.ErrNotFound
	}

	return nil
}
