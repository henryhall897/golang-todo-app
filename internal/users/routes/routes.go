package routes

import (
	"net/http"

	"github.com/henryhall897/golang-todo-app/internal/users/handler"
)

func RegisterRoutes(router *http.ServeMux, h *handler.Handler) {
	// Define user-related routes
	router.Handle("/users", handler.MethodHandler("POST", (h.CreateUserHandler)))
	router.Handle("/users/{id}", handler.MethodHandler("GET", handler.VerifyUserID(h.GetUserByIDHandler)))
	router.Handle("/users", handler.MethodHandler("GET", handler.WhichGetUsers(h.GetUsersHandler)))
	router.Handle("/users/{id}", handler.MethodHandler("PUT", handler.VerifyUserID(h.UpdateUserHandler)))
	router.Handle("/users/{id}", handler.MethodHandler("DELETE", handler.VerifyUserID(h.DeleteUserHandler)))

}
