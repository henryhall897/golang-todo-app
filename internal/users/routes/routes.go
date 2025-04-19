package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/henryhall897/golang-todo-app/internal/users/handler"
)

// RegisterRoutes sets up application routes
func RegisterRoutes(router *http.ServeMux, h *handler.Handler) {
	// Handle `/users` (List Users, Create User)
	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateUserHandler(w, r)
			return
		}

		if r.Method == http.MethodGet {
			email := r.URL.Query().Get("email")
			if email != "" {
				h.GetUserByEmailHandler(w, r)
			} else {
				h.GetUsersHandler(w, r)
			}
			return
		}

		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	router.Handle("/users/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Received request: %s %s\n", r.Method, r.URL.Path)
		// Extract path segments
		segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		fmt.Printf("length of segments: %d\n", len(segments))

		// Handle `/users/{id}`
		if len(segments) == 2 {

			if r.Method == http.MethodGet {
				handler.VerifyUserID(h.GetUserByIDHandler).ServeHTTP(w, r)
				return
			}

			if r.Method == http.MethodPut {
				handler.VerifyUserID(h.UpdateUserHandler).ServeHTTP(w, r)
				return
			}

			if r.Method == http.MethodDelete {
				handler.VerifyUserID(h.DeleteUserHandler).ServeHTTP(w, r)
				return
			}

			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Return 404 for invalid paths
		http.NotFound(w, r)
	}))

}
