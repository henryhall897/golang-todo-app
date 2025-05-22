package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/pkg/jwt/domain"
)

type TokenConfig struct {
	SecretKey     string        // Secret for signing
	TokenDuration time.Duration // e.g., 15 minutes
	Issuer        string        // e.g., "golang-todo-app"
	//Audience	  string        // e.g., "users" Todo - add audience support
}

type JWTTokenGenerator struct {
	config TokenConfig
}

func NewJWTTokenGenerator(config TokenConfig) *JWTTokenGenerator {
	return &JWTTokenGenerator{config: config}
}

func (g *JWTTokenGenerator) Gen(ctx context.Context, user domain.Payload) (string, error) {
	now := time.Now()
	exp := now.Add(g.config.TokenDuration)

	claims := domain.Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.config.Issuer,
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        uuid.New().String(), // 🔑 jti
			// Audience:  []string{g.config.Audience}, // Optional for later
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(g.config.SecretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (g *JWTTokenGenerator) Parse(ctx context.Context, tokenStr string) (domain.Claims, error) {
	var claims domain.Claims

	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(g.config.SecretKey), nil
	})

	// Token is expired → mapped to domain sentinel
	if errors.Is(err, jwt.ErrTokenExpired) {
		return domain.Claims{}, domain.ErrTokenExpired
	}

	// Any other error OR invalid token → generic invalid
	if err != nil || !token.Valid {
		return domain.Claims{}, domain.ErrTokenInvalid
	}

	return claims, nil
}
