package users

import (
	"net/http"

	"go.uber.org/zap"
)

func RegisterUserRoutes(router *http.ServeMux, usersRepo UserRepository, logger *zap.SugaredLogger) {
	userHandler := &UserHandler{
		Repo:   usersRepo,
		Logger: logger,
	}

	// Define handlers for dynamic route `/users/{id}`
	dynamicHandlers := map[string]http.HandlerFunc{
		"GET":    userHandler.GetUserByIDHandler,
		"PUT":    userHandler.UpdateUserHandler,
		"DELETE": userHandler.DeleteUserHandler,
	}

	// Define user-related routes
	router.Handle("/users", MethodHandler("POST", userHandler.CreateUserHandler)) // Create user
	router.Handle("/users/", DynamicRouteHandler(dynamicHandlers))                // Handle dynamic routes
	router.Handle("/users/email", MethodHandler("GET", userHandler.GetUserByEmailHandler))
	// router to dynamic paths and routes

}
