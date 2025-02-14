package routes

import (
	"net/http"

	"github.com/henryhall897/golang-todo-app/internal/core/logging"
	"github.com/henryhall897/golang-todo-app/internal/handlers"
	"github.com/henryhall897/golang-todo-app/internal/users"
)

func RegisterUserRoutes(router *http.ServeMux, usersStore users.Store, logger logging.Logger) {
	userHandler := &handlers.UserHandler{
		Store:  usersStore,
		Logger: logger,
	}

	// Define handlers for dynamic route `/users/{id}`
	dynamicHandlers := map[string]http.HandlerFunc{
		"GET":    userHandler.GetUserByIDHandler,
		"PUT":    userHandler.UpdateUserHandler,
		"DELETE": userHandler.DeleteUserHandler,
	}

	// Define user-related routes
	router.Handle("/users", handlers.MethodHandler("POST", userHandler.CreateUserHandler)) // Create user
	router.Handle("/users/", handlers.DynamicRouteHandler(dynamicHandlers))                // Handle dynamic routes
	router.Handle("/users/email", handlers.MethodHandler("GET", userHandler.GetUserByEmailHandler))
	// router to dynamic paths and routes

}
