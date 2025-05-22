package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/henryhall897/golang-todo-app/internal/auth/domain"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	udomain "github.com/henryhall897/golang-todo-app/internal/users/domain"
	jdomain "github.com/henryhall897/golang-todo-app/pkg/jwt/domain"
)

type service struct {
	repo     domain.Repository
	cache    domain.Cache
	uService udomain.Service
	token    jdomain.TokenGenerator
	logger   zap.SugaredLogger
}

func New(repo domain.Repository, cache domain.Cache, uService udomain.Service, token jdomain.TokenGenerator, logger *zap.SugaredLogger) domain.Service {
	return &service{
		repo:     repo,
		cache:    cache,
		uService: uService,
		token:    token,
		logger:   *logger,
	}
}

func (s *service) LoginOrRegister(ctx context.Context, input domain.AuthLoginParams) (string, udomain.User, error) {
	authidentity, err := s.repo.GetAuthIdentityByAuthID(ctx, input.AuthID)
	if err == nil {
		// Existing user → fetch and issue token
		user, _ := s.uService.GetUserByID(ctx, authidentity.UserID)
		token, err := s.token.Gen(ctx, jdomain.Payload{
			ID:   user.ID,
			Role: user.Role,
		})
		if err != nil {
			s.logger.Errorw("Failed to generate token", "error", err)
			return "", udomain.User{}, err
		}

		return token, user, nil
	}

	if !errors.Is(err, common.ErrNotFound) {
		return "", udomain.User{}, err
	}
	userRole := udomain.DefaultRole
	// New user → create in user service
	user, err := s.uService.CreateUser(ctx, udomain.CreateUserParams{
		Name:  input.Name,
		Email: input.Email,
		Role:  string(userRole),
	})
	if err != nil {
		return "", udomain.User{}, err
	}

	// Create auth identity
	_, err = s.repo.CreateAuthIdentity(ctx, domain.CreateAuthIdentityParams{
		UserID:   user.ID,
		AuthID:   input.AuthID,
		Provider: input.Provider,
	})
	if err != nil {
		return "", udomain.User{}, err
	}

	token, err := s.token.Gen(ctx, jdomain.Payload{
		ID:   user.ID,
		Role: user.Role,
	})
	if err != nil {
		s.logger.Errorw("Failed to generate token", "error", err)
		return "", udomain.User{}, err
	}
	return token, user, nil
}

func (s *service) Logout(ctx context.Context, token string) error {
	claims, err := s.token.Parse(ctx, token)
	if err != nil {
		if errors.Is(err, jdomain.ErrTokenExpired) {
			s.logger.Infof("Token expired. Skipping blacklist.")
			return nil
		}
		return fmt.Errorf("invalid token: %w", err)
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		s.logger.Infof("Token already expired by TTL. No need to blacklist.")
		return nil
	}

	err = s.cache.BlacklistToken(ctx, claims.ID, ttl)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	s.logger.Infof("Token with jti %s successfully blacklisted", claims.ID)
	return nil
}

func (s *service) ValidateToken(ctx context.Context, token string) (domain.TokenInfo, error) {
	claims, err := s.token.Parse(ctx, token)
	if err != nil {
		return domain.TokenInfo{}, fmt.Errorf("invalid token: %w", err)
	}

	blacklisted, err := s.cache.IsTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return domain.TokenInfo{}, err
	}
	if blacklisted {
		return domain.TokenInfo{}, jdomain.ErrTokenBlacklisted
	}

	return domain.TokenInfo{
		UserID: claims.UserID,
		Role:   claims.Role,
	}, nil
}
