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

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		pool: pool,
	}
}

func (s *Store) CreateUser(ctx context.Context, name, email string) (User, error) {
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

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	query := gen.New(s.pool)

	user, err := query.GetUserByID(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, common.ErrNotFound
	} else if err != nil {
		return User{}, fmt.Errorf("failed to get user by id: %w", err)
	}

	result, err := pgToUsers(user)
	if err != nil {
		return User{}, err
	}

	return result, nil
}
