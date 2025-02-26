package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"

	"go.uber.org/zap"
)

type service struct {
	repo   domain.Repository
	logger *zap.SugaredLogger
}

func NewService(repo domain.Repository, logger *zap.SugaredLogger) domain.Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) CreateUser(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
	user, err := s.repo.CreateUser(ctx, params)
	if err != nil {
		if errors.Is(err, common.ErrEmailAlreadyExists) {
			s.logger.Warnw("CreateUser failed: email already exists",
				"email", params.Email,
			)
			return domain.User{}, common.ErrEmailAlreadyExists
		}

		s.logger.Errorw("Failed to create user",
			"error", err,
			"name", params.Name,
			"email", params.Email,
		)
		return domain.User{}, common.ErrInternalServerError
	}

	s.logger.Infow("User created successfully",
		"user_id", user.ID,
		"name", user.Name,
	)
	return user, nil
}

// TODO 1: Implement the GetUserByID method
func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

// TODO 2: Implement the GetUserByEmail method
func (s *service) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

// TODO 3: Implement the ListUsers method
func (s *service) ListUsers(ctx context.Context, params domain.ListUsersParams) ([]domain.User, error) {
	return s.repo.ListUsers(ctx, params)
}

// TODO 4: Implement the UpdateUser method
func (s *service) UpdateUser(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
	return s.repo.UpdateUser(ctx, params)
}

// TODO 5: Implement the DeleteUser method
func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteUser(ctx, id)
}
