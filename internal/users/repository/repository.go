package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/henryhall897/golang-todo-app/gen/queries/userstore"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	pool  *pgxpool.Pool
	query *userstore.Queries
}

func New(pool *pgxpool.Pool) *repository {
	return &repository{
		pool:  pool,
		query: userstore.New(pool),
	}
}

func (r *repository) CreateUser(ctx context.Context, newUser domain.CreateUserParams) (domain.User, error) {
	// Convert the CreateUserParams to gen.CreateUserParams
	pgNewUser := createUserParamsToPG(newUser)

	// Execute the query to create a new user
	user, err := r.query.CreateUser(ctx, pgNewUser)
	if err != nil {
		// Check if the error is a unique constraint violation
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // 23505 is PostgreSQL's unique violation error code
			return domain.User{}, fmt.Errorf("%w", ErrEmailAlreadyExists)
		}
		return domain.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	result, err := pgToUsers(user)
	if err != nil {
		return domain.User{}, err
	}

	return result, nil
}

func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	//convert uuid.UUID to pgtype.UUID. not checking for error because handler verified the UUID
	pgUUID, _ := common.ToPgUUID(id)

	// Execute the query to get the user by ID
	user, err := r.query.GetUserByID(ctx, pgUUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, fmt.Errorf("user %s: %w", id, common.ErrNotFound)
	} else if err != nil {
		return domain.User{}, fmt.Errorf("user %s: %w", id, common.ErrInternalServerError)

	}

	// Convert the raw database results into the domain.User type
	result, err := pgToUsers(user)
	if err != nil {
		return domain.User{}, err
	}

	return result, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	// Execute the query to get the user by email
	user, err := r.query.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, fmt.Errorf("email %s: %w", email, common.ErrNotFound)
	} else if err != nil {
		return domain.User{}, fmt.Errorf("email %s: %w", email, common.ErrInternalServerError)
	}

	// Convert the raw database results into the domain.User type
	result, err := pgToUsers(user)
	if err != nil {
		return domain.User{}, err
	}

	return result, nil
}

func (r *repository) GetUsers(ctx context.Context, getUserParams domain.GetUsersParams) ([]domain.User, error) {
	// Execute the query to get users
	users, err := r.query.GetUsers(ctx, getUsersParamsToPG(getUserParams))
	if errors.Is(err, pgx.ErrNoRows) {
		return []domain.User{}, fmt.Errorf("no users found: %w", common.ErrNotFound)
	} else if err != nil {
		return []domain.User{}, fmt.Errorf("failed to list users: %w", common.ErrInternalServerError)
	}

	// Convert the raw database results into the domain.User type
	results := make([]domain.User, 0, len(users))
	for _, u := range users {
		result, err := pgToUsers(u)
		if err != nil {
			return []domain.User{}, fmt.Errorf("failed to convert user: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *repository) UpdateUser(ctx context.Context, updateParams domain.UpdateUserParams) (domain.User, error) {
	// Transform input to the required database structure Handler checks for valid UUID. can ignore error here
	arg, _ := updateUserParamsToPG(updateParams)

	// Execute the update query
	dbUpdatedUser, err := r.query.UpdateUser(ctx, arg)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, common.ErrNotFound
	} else if err != nil {
		return domain.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	// Convert the updated user to the application-level model
	updatedUser, err := pgToUsers(dbUpdatedUser)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to convert updated user: %w", err)
	}

	return updatedUser, nil
}

func (r *repository) DeleteUser(ctx context.Context, id uuid.UUID) error {

	// Convert uuid.UUID to pgtype.UUID - Handler checks for valid UUID. can ignore error here
	pgId, _ := common.ToPgUUID(id)

	// Execute the delete query
	rowsAffected, err := r.query.DeleteUser(ctx, pgId)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}

	if rowsAffected == 0 {
		return common.ErrNotFound
	}

	return nil
}
