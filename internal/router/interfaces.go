package router

import (
	"net/http"

	"github.com/henryhall897/golang-todo-app/internal/users/handler"
)

// RouteRegisterFunc.
type RouteRegisterFunc func(router *http.ServeMux, h *handler.Handler)
