package domain

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenClaims represents JWT claims issued to a user
type Claims struct {
	UserID uuid.UUID `json:"user_id"` // custom claim
	Role   string    `json:"role"`    // custom claim

	jwt.RegisteredClaims // includes standard fields like exp, iat, jti, iss, sub
}

type Payload struct {
	ID   uuid.UUID `json:"id"`
	Role string    `json:"role"`
}
