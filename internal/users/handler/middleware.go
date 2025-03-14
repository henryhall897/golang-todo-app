package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/core/logging"
)

type contextKey string

const userIDKey = contextKey("userID")

// VerifyUserID extracts and validates a UUID from the request context
func VerifyUserID(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve logger from context
		logger := logging.GetLogger(r.Context())

		// Extract user ID from URL path
		segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(segments) < 2 || segments[1] == "" {
			http.NotFound(w, r)
			return
		}
		userIDStr := segments[1]

		// Convert userID string to UUID
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			logger.Warnw("VerifyUserID failed: invalid user ID format", "user_id", userIDStr, "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Ensure UUID is not nil
		if userID == uuid.Nil {
			logger.Warnw("VerifyUserID failed: nil UUID provided", "user_id", userIDStr)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		logger.Infow("VerifyUserID successfully validated user ID", "user_id", userID)

		// Store validated UUID in request context and proceed
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		logger.Infow("VerifyUserID successfully added user_id to the context userID", "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
