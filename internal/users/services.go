package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/core/common"
	"go.uber.org/zap"
)

type UserService interface {
	CreateUser(ctx context.Context, params CreateUserParams) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	ListUsers(ctx context.Context, params ListUsersParams) ([]User, error)
	UpdateUser(ctx context.Context, params UpdateUserParams) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo   Repository
	logger *zap.SugaredLogger
}

/*func NewUserService(repo Repository, logger *zap.SugaredLogger) Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}*/

func (s *Service) CreateUser(ctx context.Context, params CreateUserParams) (User, error) {
	if params.Name == "" || params.Email == "" {
		s.logger.Warnw("CreateUser failed: missing required fields",
			"provided_name", params.Name,
			"provided_email", params.Email,
		)
		return User{}, common.ErrInvalidRequestBody
	}

	user, err := s.repo.CreateUser(ctx, params)
	if err != nil {
		if errors.Is(err, common.ErrEmailAlreadyExists) {
			s.logger.Warnw("CreateUser failed: email already exists",
				"email", params.Email,
			)
			return User{}, common.ErrEmailAlreadyExists
		}

		s.logger.Errorw("Failed to create user",
			"error", err,
			"name", params.Name,
			"email", params.Email,
		)
		return User{}, errors.New("internal server error")
	}

	s.logger.Infow("User created successfully",
		"user_id", user.ID,
		"name", user.Name,
	)
	return user, nil
}
