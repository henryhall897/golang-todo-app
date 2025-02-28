package routes

import (
	"net/http"

	"github.com/henryhall897/golang-todo-app/internal/users/handler"
)

type UserRoutes struct {
	Handler *handler.Handler
}

// newUserRoutes creates a new UserRoutes instance.
func NewUserRoutes(handler *handler.Handler) *UserRoutes {
	return &UserRoutes{Handler: handler}
}

func (u *UserRoutes) RegisterRoutes(router *http.ServeMux) {
	// Define user-related routes
	router.Handle("/users", handler.MethodHandler("POST", handler.VerifyCreateUserBody(u.Handler.CreateUserHandler())))
	router.Handle("/users/{id}", handler.MethodHandler("GET", handler.VerifyUserID(u.Handler.GetUserByIDHandler())))
	router.Handle("/users", handler.MethodHandler("GET", handler.VerifyGetUsersQuery(u.Handler.GetUsersHandler())))

}
