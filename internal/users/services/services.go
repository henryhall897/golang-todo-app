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
	cache  domain.Cache
	logger *zap.SugaredLogger
}

func New(repo domain.Repository, cache domain.Cache, logger *zap.SugaredLogger) domain.Service {
	return &service{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// CreateUser creates a new user and caches it in Redis
func (s *service) CreateUser(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
	// Attempt to create user in the database
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
		return domain.User{}, err
	}

	s.logger.Infow("User created successfully",
		"user_id", user.ID,
		"name", user.Name,
	)

	// Attempt to cache the new user in Redis
	if err := s.cache.CacheUser(ctx, user); err != nil {
		s.logger.Warnw("Failed to cache user in Redis",
			"user_id", user.ID,
			"error", err,
		)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID, using Redis caching when possible
func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	// Attempt to retrieve from cache
	if cachedUser, err := s.cache.GetUserByID(ctx, id); err == nil {
		s.logger.Debugw("Cache hit: Retrieved user from Redis", "user_id", cachedUser.ID)
		return cachedUser, nil
	}

	// Cache miss or deserialization failure, proceed to DB lookup
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		// Handle repository-level errors
		if errors.Is(err, repository.ErrInvalidDbUserID) || errors.Is(err, repository.ErrFailedToParseUUID) {
			s.logger.Errorw("GetUserByID failed: Invalid user data in database",
				"user_id", id.String(),
				"error", err,
			)
			return domain.User{}, common.ErrInternalServerError // Mask as internal error
		}

		if errors.Is(err, common.ErrNotFound) {
			s.logger.Warnw("GetUserByID failed: user not found", "user_id", id.String())
			return domain.User{}, common.ErrNotFound
		}

		// Log unexpected errors
		s.logger.Errorw("GetUserByID failed: unexpected error",
			"user_id", id.String(),
			"error", err,
		)
		return domain.User{}, common.ErrInternalServerError
	}

	s.logger.Debugw("User retrieved successfully", "user_id", user.ID, "name", user.Name)

	// Store retrieved user in cache with an expiration
	if err := s.cache.CacheUser(ctx, user); err != nil {
		s.logger.Warnw("Failed to cache user in Redis", "user_id", user.ID, "error", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email, utilizing Redis caching
func (s *service) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	// Attempt to retrieve cached user from Redis
	if cachedUser, err := s.cache.GetUserByEmail(ctx, email); err == nil {
		s.logger.Debugw("Cache hit: Retrieved user from Redis",
			"user_id", cachedUser.ID,
			"email", cachedUser.Email,
		)
		return cachedUser, nil
	}

	// Fetch user from the database
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidDbUserID) || errors.Is(err, repository.ErrFailedToParseUUID) {
			s.logger.Errorw("GetUserByEmail failed: invalid user data in database",
				"email", email, "error", err,
			)
			return domain.User{}, common.ErrInternalServerError
		}
		if errors.Is(err, common.ErrNotFound) {
			s.logger.Warnw("GetUserByEmail failed: user not found",
				"email", email, "error", err,
			)
			return domain.User{}, common.ErrNotFound
		}
		s.logger.Errorw("GetUserByEmail failed: internal server error",
			"email", email, "error", err,
		)
		return domain.User{}, common.ErrInternalServerError
	}

	// Store the retrieved user in Redis for future lookups
	if err := s.cache.CacheUser(ctx, user); err != nil {
		s.logger.Errorw("Failed to store user in Redis", "email", email, "error", err)
	}

	s.logger.Debugw("User retrieved successfully", "user_id", user.ID, "email", user.Email)
	return user, nil
}

/*// GetUserByAuthID retrieves a user by their Auth0 ID, utilizing Redis caching
func (s *service) GetUserByAuthID(ctx context.Context, authID string) (domain.User, error) {
	// Attempt to retrieve cached user from Redis
	if cachedUser, err := s.cache.GetUserByAuthID(ctx, authID); err == nil {
		s.logger.Debugw("Cache hit: Retrieved user from Redis via Auth0 ID",
			"user_id", cachedUser.ID,
			"auth_id", authID,
		)
		return cachedUser, nil
	}

	// Fetch user from the database
	user, err := s.repo.GetUserByAuthID(ctx, authID)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidDbUserID) || errors.Is(err, repository.ErrFailedToParseUUID) {
			s.logger.Errorw("GetUserByAuthID failed: invalid user data in database",
				"auth_id", authID, "error", err,
			)
			return domain.User{}, common.ErrInternalServerError
		}
		if errors.Is(err, common.ErrNotFound) {
			s.logger.Warnw("GetUserByAuthID failed: user not found",
				"auth_id", authID, "error", err,
			)
			return domain.User{}, common.ErrNotFound
		}
		s.logger.Errorw("GetUserByAuthID failed: internal server error",
			"auth_id", authID, "error", err,
		)
		return domain.User{}, common.ErrInternalServerError
	}

	// cache to redis Redis
	if err := s.cache.CacheUser(ctx, user); err != nil {
		s.logger.Errorw("Failed to store user in Redis", "auth_id", authID, "error", err)
	}

	s.logger.Debugw("User retrieved successfully", "user_id", user.ID, "auth_id", user.AuthID)
	return user, nil
}*/

// GetUsers retrieves a list of users with caching
func (s *service) GetUsers(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
	// Attempt to retrieve cached users from Redis
	if cachedUsers, err := s.cache.GetUserByPagination(ctx, params); err == nil {
		s.logger.Debugw("Cache hit: Retrieved users from Redis",
			"user_count", len(cachedUsers),
			"params", params,
		)
		return cachedUsers, nil
	}

	// Fetch users from the database
	users, err := s.repo.GetUsers(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidDbUserID) || errors.Is(err, repository.ErrFailedToParseUUID) {
			s.logger.Errorw("GetUsers failed: invalid user data in database",
				"params", params, "error", err,
			)
			return nil, common.ErrInternalServerError
		}
		if errors.Is(err, common.ErrNotFound) {
			s.logger.Warnw("GetUsers failed: no users found",
				"params", params, "error", err,
			)
			return nil, common.ErrNotFound
		}
		s.logger.Errorw("GetUsers failed: internal server error",
			"params", params, "error", err,
		)
		return nil, common.ErrInternalServerError
	}

	// Store the retrieved users in Redis for future queries
	if err := s.cache.CacheUserByPagination(ctx, users, params); err != nil {
		s.logger.Errorw("Failed to store users in Redis", "params", params, "error", err)
	}

	s.logger.Debugw("Users retrieved successfully", "user_count", len(users), "params", params)
	return users, nil
}

// TODO - Implement AUTH0 update
// UpdateUser updates an existing user's details and refreshes cache
func (s *service) UpdateUser(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
	// Clear out old cache entries
	s.clearUserCache(ctx, params.ID)

	// Update user in the database
	user, err := s.repo.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return domain.User{}, common.ErrNotFound
		} else if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return domain.User{}, ErrEmailAlreadyExists
		}
		s.logger.Errorw("UpdateUser failed: unexpected internal error",
			"user_id", params.ID,
			"error", err,
		)
		return domain.User{}, common.ErrInternalServerError
	}

	// Store updated user in Redis
	if err := s.cache.CacheUser(ctx, user); err != nil {
		s.logger.Warnw("Failed to store updated user in Redis", "user_id", user.ID, "error", err)
	}

	s.logger.Infow("User updated successfully and cache refreshed",
		"user_id", user.ID,
		"updated_fields", params,
	)
	return user, nil
}

// TODO - Implement AUTH0 deletion
// DeleteUser deletes a user by ID
func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Clear out old cache entries
	s.clearUserCache(ctx, id)

	// Attempt to delete user from the database
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
