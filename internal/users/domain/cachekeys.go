package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// CacheKeyByID generates a cache key for a user by their ID.
func CacheKeyByID(id uuid.UUID) string {
	return id.String()
}

// CacheKeyByEmail generates a cache key for a user by their email.
func CacheKeyByEmail(email string) string {
	return fmt.Sprintf("%s:%s", RedisEmailPrefix, email)
}

// CacheKeyByPagination generates a cache key for paginated user results.
func CacheKeyByPagination(limit, offset int) string {
	return fmt.Sprintf("page:limit=%d:offset=%d", limit, offset)
}

// CacheKeyByAuthID generates a cache key for a user by their AuthID.
func CacheKeyByAuthID(authID string) string {
	return fmt.Sprintf("%s:%s", RedisAuthIDPrefix, authID)
}
