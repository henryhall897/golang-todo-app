package domain

import (
	"context"

	udomain "github.com/henryhall897/golang-todo-app/internal/users/domain"
)

//go:generate moq -out=../../../gen/mocks/jwtmock/token_generator_mock.go -pkg=jwtmock . TokenGenerator
type TokenGenerator interface {
	GenerateToken(ctx context.Context, user udomain.User) (string, error)
	ParseToken(ctx context.Context, tokenString string) (TokenClaims, error)
}
