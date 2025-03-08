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
		// Handle repository-level errors that indicate database issues
		if errors.Is(err, repository.ErrInvalidDbUserID) || errors.Is(err, repository.ErrFailedToParseUUID) {
			s.logger.Errorw("GetUserByID failed: Invalid user data in database",
				"user_id", id.String(),
				"error", err,
			)
			return domain.User{}, common.ErrInternalServerError // Mask as internal error
		}

		if errors.Is(err, common.ErrNotFound) {
			s.logger.Warnw("GetUserByID failed: user %s not found", id.String())
			return domain.User{}, common.ErrNotFound
		}

		// Log any unexpected errors
		s.logger.Errorw("GetUserByID failed: unexpected error",
			"user_id", id.String(),
			"error", err,
		)
		return domain.User{}, common.ErrInternalServerError
	}

	s.logger.Debugw("User retrieved successfully",
		"user_id", user.ID,
		"name", user.Name,
	)
	return user, nil
}

// GetUsers retrieves a list of users
func (s *service) GetUsers(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
	users, err := s.repo.GetUsers(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidDbUserID) || errors.Is(err, repository.ErrFailedToParseUUID) {
			// Log and mask database corruption issues as internal errors
			s.logger.Errorw("GetUsers failed: user data is invalid in the database",
				"params", params,
				"error", err,
			)
			return []domain.User{}, common.ErrInternalServerError
		}

		if errors.Is(err, common.ErrNotFound) {
			s.logger.Warnw("GetUsers failed: no users found",
				"params", params,
				"error", err,
			)
			return []domain.User{}, common.ErrNotFound
		}

		// Log unexpected errors
		s.logger.Errorw("GetUsers failed: internal server error",
			"params", params,
			"error", err,
		)
		return []domain.User{}, common.ErrInternalServerError
	}

	s.logger.Debugw("Users retrieved successfully",
		"user_count", len(users),
		"params", params,
	)
	return users, nil
}

// GetUserByEmail retrieves a user by email
func (s *service) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidDbUserID) || errors.Is(err, repository.ErrFailedToParseUUID) {
			// Log and mask database corruption issues as internal errors
			s.logger.Errorw("GetUserByEmail failed: user data is invalid in the database",
				"email", email,
				"error", err,
			)
			return domain.User{}, common.ErrInternalServerError
		}

		if errors.Is(err, common.ErrNotFound) {
			s.logger.Warnw("GetUserByEmail failed: user %s not found", email)
			return domain.User{}, common.ErrNotFound
		}

		// Log unexpected errors
		s.logger.Errorw("GetUserByEmail failed: internal server error",
			"email", email,
			"error", err,
		)
		return domain.User{}, common.ErrInternalServerError
	}

	s.logger.Infow("User retrieved successfully",
		"user_id", user.ID,
		"email", user.Email,
	)
	return user, nil
}

// UpdateUser updates an existing user's details
func (s *service) UpdateUser(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
	user, err := s.repo.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			// User not found, return meaningful error
			return domain.User{}, common.ErrNotFound
		}

		if errors.Is(err, repository.ErrInvalidDbUserID) || errors.Is(err, repository.ErrFailedToParseUUID) {
			// Log and mask database corruption issues as internal errors
			s.logger.Errorw("UpdateUser failed: invalid user data in database",
				"user_id", params.ID,
				"error", err,
			)
			return domain.User{}, common.ErrInternalServerError
		}

		// Log unexpected errors
		s.logger.Errorw("UpdateUser failed: unexpected internal error",
			"user_id", params.ID,
			"error", err,
		)
		return domain.User{}, common.ErrInternalServerError
	}

	s.logger.Infow("User updated successfully",
		"user_id", user.ID,
		"updated_fields", params, // Log updated fields
	)
	return user, nil
}

// DeleteUser deletes a user by ID
func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return common.ErrNotFound
		}

		// Log unexpected errors before returning an internal server error
		s.logger.Errorw("DeleteUser failed: internal server error",
			"user_id", id,
			"error", err,
		)
		return common.ErrInternalServerError
	}
	return nil
}
