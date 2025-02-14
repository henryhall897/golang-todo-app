package handlers

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey = contextKey("userID")

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

func DynamicRouteHandler(handlers map[string]http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the path segments (e.g., /users/{id})
		pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathSegments) != 2 || pathSegments[0] != "users" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// Extract the dynamic path variable (e.g., user ID)
		userID := pathSegments[1]
		ctx := context.WithValue(r.Context(), userIDKey, userID)

		// Check if there's a handler for the request method
		handler, exists := handlers[r.Method]
		if !exists {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Call the handler with the updated context
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
