package testutils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
)

// GenerateMockUsers creates a specified number of mock users with unique emails.
func GenerateMockUsers(count int) []domain.User {
	now := time.Now()
	userList := make([]domain.User, count)
	for i := 0; i < count; i++ {
		userList[i] = domain.User{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("John %d Doe", i+1),
			Email:     fmt.Sprintf("johndoe%d@example.com", i+1),
			CreatedAt: &now,
			UpdatedAt: &now,
		}
	}
	return userList
}
