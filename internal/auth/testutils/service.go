package testutils

import (
	"fmt"
	"time"

	domain "github.com/henryhall897/golang-todo-app/internal/auth/domain"
	userdomain "github.com/henryhall897/golang-todo-app/internal/users/domain"
)

const (
	// MockProvider is a mock provider name for testing.
	MockProviderBase = "mock_provider"
)

// service level
// GenerateMockAuthIdentities creates mock auth identity records for a list of users.
func GenerateMockAuthIdentities(users []userdomain.User, provider string) []domain.AuthIdentity {
	now := time.Now()
	authList := make([]domain.AuthIdentity, len(users))

	for i, user := range users {
		authList[i] = domain.AuthIdentity{
			AuthID:    fmt.Sprintf("%s|mock%d", provider, i+1),
			Provider:  provider,
			UserID:    user.ID,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	return authList
}

// GenerateMockAuthIdentitiesForUser generates a list of unique auth identities for a single user,
// each with a unique provider (e.g., mock_provider_1, mock_provider_2).
func GenerateMockAuthIdentitiesForUser(user userdomain.User, count int) []domain.AuthIdentity {
	now := time.Now()
	auths := make([]domain.AuthIdentity, count)

	for i := range count {
		provider := fmt.Sprintf("%s_%d", MockProviderBase, i+1)
		auths[i] = domain.AuthIdentity{
			AuthID:    fmt.Sprintf("%s|mock%d", provider, i+1),
			Provider:  provider,
			UserID:    user.ID,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	return auths
}
