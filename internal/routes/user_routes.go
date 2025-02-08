package routes

import (
	"golang-todo-app/internal/core/logging"
	"golang-todo-app/internal/handlers"
	"golang-todo-app/internal/users"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(router *mux.Router, usersStore users.Store, logger logging.Logger) {
	userHandler := &handlers.UserHandler{
		Store:  usersStore,
		Logger: logger,
	}

	// Define user-related routes
	router.HandleFunc("/users", userHandler.CreateUserHandler).Methods("POST")
	router.HandleFunc("/users/{id}", userHandler.GetUserByIDHandler).Methods("GET")
	router.HandleFunc("/users", userHandler.ListUsersHandler).Methods("GET")
	router.HandleFunc("/users/email", userHandler.GetUserByEmailHandler).Methods("GET")
	router.HandleFunc("/users/{id}", userHandler.UpdateUserHandler).Methods("PUT")
	router.HandleFunc("/users/{id}", userHandler.DeleteUserHandler).Methods("DELETE")
}
