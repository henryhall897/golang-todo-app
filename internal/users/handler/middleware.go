package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/core/logging"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
)

type contextKey string

const userIDKey = contextKey("userID")
const queryParamsKey = contextKey("queryParams")

// methodHandler filters requests by HTTP method
func MethodHandler(method string, handlerFunc http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		handlerFunc(w, r)
	})
}

// VerifyUserID extracts and validates a UUID from the request context
func VerifyUserID(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve logger from context
		logger := logging.GetLogger(r.Context())

		// Extract user ID from the URL path using standard library
		pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathSegments) < 2 {
			logger.Warnw("VerifyUserID failed: user ID missing in path")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		userIDStr := pathSegments[1] // Second segment is the {id}

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

		// Store validated user ID in request context and proceed
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WhichGetUsers extracts and validates query parameters for user retrieval
func WhichGetUsers(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger(r.Context())

		// Extract query parameters
		email := r.URL.Query().Get("email")
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		// Default query type and pagination values
		queryType := domain.QueryTypeList
		limit := domain.DefaultLimit
		offset := domain.DefaultOffset

		// Validate email if provided
		if email != "" {
			queryType = domain.QueryTypeEmail
			if len(email) < 3 || len(email) > 320 {
				logger.Warnw("WhichGetUsers failed: invalid email query parameter", "email", email)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		} else {
			// Validate limit if provided
			if limitStr != "" {
				parsedLimit, err := strconv.Atoi(limitStr)
				if err != nil || parsedLimit <= 0 {
					logger.Warnw("WhichGetUsers failed: invalid limit parameter", "limit", limitStr)
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				}
				limit = parsedLimit
			}

			// Validate offset if provided
			if offsetStr != "" {
				parsedOffset, err := strconv.Atoi(offsetStr)
				if err != nil || parsedOffset < 0 {
					logger.Warnw("WhichGetUsers failed: invalid offset parameter", "offset", offsetStr)
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					return
				}
				offset = parsedOffset
			}
		}

		// Store query parameters in context
		queryParams := domain.GetQueryParams{
			QueryType: queryType,
			Email:     email,
			Limit:     limit,
			Offset:    offset,
		}
		ctx := context.WithValue(r.Context(), queryParamsKey, queryParams)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
