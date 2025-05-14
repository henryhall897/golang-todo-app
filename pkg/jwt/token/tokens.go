package token

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	udomain "github.com/henryhall897/golang-todo-app/internal/users/domain"
)

type TokenConfig struct {
	SecretKey     string        // Secret for signing
	TokenDuration time.Duration // e.g., 15 minutes
	Issuer        string        // e.g., "golang-todo-app"
}

type JWTTokenGenerator struct {
	config TokenConfig
}

func NewJWTTokenGenerator(config TokenConfig) *JWTTokenGenerator {
	return &JWTTokenGenerator{config: config}
}

// GenerateToken generates a signed JWT for a given user and auth identity.
func (g *JWTTokenGenerator) GenerateToken(ctx context.Context, user udomain.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": now.Add(g.config.TokenDuration).Unix(),
		"iat": now.Unix(),
		"iss": g.config.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(g.config.SecretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
