package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
)

// clearUserCache fetches a user by ID and removes them from cache
func (s *service) clearUserCache(ctx context.Context, id uuid.UUID) {
	// Try to get user from cache first
	user, err := s.cache.GetUserByID(ctx, id)
	if err != nil {
		// If cache miss, try to get from database
		dbUser, err := s.repo.GetUserByID(ctx, id)
		if err == nil {
			// If found in DB, use that for cache clearing
			user = dbUser
			s.cache.DeleteUserByID(ctx, user.ID)
			s.cache.DeleteUserByEmail(ctx, user.Email)
		} else {
			// Just attempt to delete by ID if we can't get the user
			s.cache.DeleteUserByID(ctx, id)
			s.logger.Warnw("Could not retrieve full user data for cache clearing",
				"user_id", id,
				"error", err)
		}
		return
	}

	// If we have the user, delete both cache entries
	s.cache.DeleteUserByID(ctx, user.ID)
	s.cache.DeleteUserByEmail(ctx, user.Email)
}

// IsValidRole returns true if the role is part of AllRoles
func IsValidRole(role domain.Role) bool {
	_, ok := domain.AllRoles[role]
	return ok
}
