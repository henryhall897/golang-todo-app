package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuthIdentity struct {
	AuthID    string    `json:"auth_id"`
	Provider  string    `json:"provider"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateAuthIdentityParams struct {
	AuthID   string
	Provider string
	UserID   uuid.UUID
	Role     string
}

type UpdateAuthIdentityParams struct {
	AuthID string
	Role   string
}
