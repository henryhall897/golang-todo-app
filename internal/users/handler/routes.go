package handler

import (
	"net/http"
)

type UserRoutes struct {
	Handler *Handler
}

// newUserRoutes creates a new UserRoutes instance.
func NewUserRoutes(handler *Handler) *UserRoutes {
	return &UserRoutes{Handler: handler}
}

func (u *UserRoutes) RegisterRoutes(router *http.ServeMux) {
	// Define handlers for dynamic route `/users/{id}`
	dynamicHandlers := map[string]http.HandlerFunc{
		"GET":    u.Handler.CreateUserHandler,
		"PUT":    u.Handler.UpdateUserHandler,
		"DELETE": u.Handler.DeleteUserHandler,
	}

	// Define user-related routes
	router.Handle("/users", MethodHandler("POST", u.Handler.CreateUserHandler)) // Create user
	router.Handle("/users/", DynamicRouteHandler(dynamicHandlers))              // Handle dynamic routes
	router.Handle("/users/email", MethodHandler("GET", u.Handler.GetUserByEmailHandler))
}
