package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuthIdentity struct {
	AuthID    string    `json:"auth_id"`
	Provider  string    `json:"provider"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateAuthIdentityParams struct {
	AuthID   string    `json:"auth_id"`
	Provider string    `json:"provider"`
	UserID   uuid.UUID `json:"user_id"`
}

// DeleteAuthIdentityParams represents the parameters for deleting an auth identity.
type DeleteAuthIdentityParams struct {
	AuthID string    `json:"auth_id"`
	UserID uuid.UUID `json:"user_id"`
}

type AuthLoginParams struct {
	AuthID   string `json:"auth_id"`
	Provider string `json:"provider"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

type TokenInfo struct {
	UserID uuid.UUID
	Role   string
}
