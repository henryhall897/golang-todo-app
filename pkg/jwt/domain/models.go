package domain

import (
	"time"

	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	Issuer    string    `json:"issuer"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
