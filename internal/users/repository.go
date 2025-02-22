package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users/gen"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository defines the methods required for user operations.
type UserRepository interface {
	CreateUser(ctx context.Context, newUserParams CreateUserParams) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	ListUsers(ctx context.Context, params ListUsersParams) ([]User, error)
	UpdateUser(ctx context.Context, updateUserparams UpdateUserParams) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (s *Repository) CreateUser(ctx context.Context, newUser CreateUserParams) (User, error) {
	query := gen.New(s.pool)

	// Convert the CreateUserParams to gen.CreateUserParams
	pgNewUser := toDBCreateUserParams(newUser)

	// Execute the query to create a new user
	user, err := query.CreateUser(ctx, pgNewUser)
	if err != nil {
		// Check if the error is a unique constraint violation
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // 23505 is PostgreSQL's unique violation error code
			return User{}, fmt.Errorf("%w", common.ErrEmailAlreadyExists)
		}
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}

	result, err := pgToUsers(user)
	if err != nil {
		return User{}, err
	}

	return result, nil
}

func (s *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	query := gen.New(s.pool)

	user, err := query.GetUserByID(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, common.ErrNotFound
	} else if err != nil {
		return User{}, fmt.Errorf("user %s: %w", id, common.ErrNotFound)

	}

	result, err := pgToUsers(user)
	if err != nil {
		return User{}, err
	}

	return result, nil
}

func (s *Repository) GetUserByEmail(ctx context.Context, email string) (User, error) {
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

func (s *Repository) ListUsers(ctx context.Context, listParams ListUsersParams) ([]User, error) {
	query := gen.New(s.pool)
	// Convert the ListUsersParams to gen.ListUsersParams
	pgParams := toDBListParams(listParams)

	// Execute the query to get users
	users, err := query.ListUsers(ctx, pgParams)
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

func (s *Repository) UpdateUser(ctx context.Context, updateParams UpdateUserParams) (User, error) {
	query := gen.New(s.pool)

	// Transform input to the required database structure
	arg, err := toDBUpdateUserParams(updateParams)
	if err != nil {
		return User{}, fmt.Errorf("failed to transform update parameters: %w", err)
	}

	// Execute the update query
	dbUpdatedUser, err := query.UpdateUser(ctx, arg)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, common.ErrNotFound
	} else if err != nil {
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}

	// Convert the updated user to the application-level model
	updatedUser, err := pgToUsers(dbUpdatedUser)
	if err != nil {
		return User{}, fmt.Errorf("failed to convert updated user: %w", err)
	}

	return updatedUser, nil
}

func (s *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := gen.New(s.pool)

	// Convert uuid.UUID to pgtype.UUID
	pgId, err := common.ToPgUUID(id)
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
