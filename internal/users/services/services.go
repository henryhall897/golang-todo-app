package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
	"github.com/henryhall897/golang-todo-app/internal/users/repository"

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

// CreateUser service creates a new user
func (s *service) CreateUser(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
	user, err := s.repo.CreateUser(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			s.logger.Warnw("CreateUser failed: email already exists",
				"email", params.Email,
			)
			return domain.User{}, ErrEmailAlreadyExists
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

// GetUserByID service retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		switch {
		// UUID validation should have happened in the handler, so reaching here is an unexpected failure
		case errors.Is(err, common.ErrInvalidUUID):
			s.logger.Errorw("GetUserByID failed: UUID validation was skipped or handler validation failed",
				"user_id", id.String(),
				"error", err,
			)
			return domain.User{}, common.ErrInternalServerError // Mask as internal error

		// Data retrieved from the database is invalid (corruption issue)
		case errors.Is(err, repository.ErrInvalidDbUserID), errors.Is(err, repository.ErrFailedToParseUUID):
			s.logger.Errorw("GetUserByID failed: user data is invalid in the database",
				"user_id", id.String(),
				"error", err,
			)
			return domain.User{}, common.ErrInternalServerError // Mask as internal error

		// User not found
		case errors.Is(err, common.ErrNotFound):
			s.logger.Warnw("GetUserByID failed: user not found",
				"user_id", id.String(),
			)
			return domain.User{}, common.ErrNotFound

		// Any other unexpected errors
		default:
			s.logger.Errorw("GetUserByID failed: internal server error",
				"user_id", id.String(),
				"error", err,
			)
			return domain.User{}, common.ErrInternalServerError
		}
	}

	s.logger.Infow("User retrieved successfully",
		"user_id", user.ID,
		"name", user.Name,
	)
	return user, nil
}

// GetUsers service retrieves a list of users
func (s *service) GetUsers(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
	users, err := s.repo.GetUsers(ctx, params)
	if err != nil {
		switch {
		// Data retrieved from the database is invalid (corruption issue)
		case errors.Is(err, repository.ErrInvalidDbUserID), errors.Is(err, repository.ErrFailedToParseUUID):
			s.logger.Errorw("GetUsers failed: user data is invalid in the database",
				"params", params,
				"error", err,
			)
			return []domain.User{}, common.ErrInternalServerError // Mask as internal error

		// No users found
		case errors.Is(err, common.ErrNotFound):
			s.logger.Warnw("GetUsers failed: no users found",
				"params", params,
			)
			return []domain.User{}, common.ErrNotFound

		// Any other unexpected errors
		default:
			s.logger.Errorw("GetUsers failed: internal server error",
				"params", params,
				"error", err,
			)
			return []domain.User{}, common.ErrInternalServerError
		}
	}

	s.logger.Infow("Users retrieved successfully",
		"user_count", len(users),
		"params", params,
	)
	return users, nil
}

// GetUserByEmail service retrieves a user by email
func (s *service) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		switch {
		// Data retrieved from the database is invalid (corruption issue)
		case errors.Is(err, repository.ErrInvalidDbUserID), errors.Is(err, repository.ErrFailedToParseUUID):
			s.logger.Errorw("GetUserByEmail failed: user data is invalid in the database",
				"email", email,
				"error", err,
			)
			return domain.User{}, common.ErrInternalServerError // Mask as internal error

		// User not found
		case errors.Is(err, common.ErrNotFound):
			s.logger.Warnw("GetUserByEmail failed: user not found",
				"email", email,
			)
			return domain.User{}, common.ErrNotFound

		// Any other unexpected errors
		default:
			s.logger.Errorw("GetUserByEmail failed: internal server error",
				"email", email,
				"error", err,
			)
			return domain.User{}, common.ErrInternalServerError
		}
	}

	s.logger.Infow("User retrieved successfully",
		"user_id", user.ID,
		"email", user.Email,
	)
	return user, nil
}

// TODO 4: Implement the UpdateUser method
func (s *service) UpdateUser(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
	return s.repo.UpdateUser(ctx, params)
}

// TODO 5: Implement the DeleteUser method
func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteUser(ctx, id)
}
