package domain

import (
	"context"
)

//go:generate moq -out=../../../gen/mocks/jwtmock/token_generator_mock.go -pkg=jwtmock . TokenGenerator
type TokenGenerator interface {
	Gen(ctx context.Context, user Payload) (string, error)
	Parse(ctx context.Context, tokenString string) (Claims, error)
}
