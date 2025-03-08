package router

import (
	"net/http"

	"github.com/henryhall897/golang-todo-app/internal/middleware"
	"github.com/henryhall897/golang-todo-app/internal/users/handler"
)

// router manages the routes for the application.
type Router struct {
	Mux            *http.ServeMux
	LimitedHandler http.Handler
}

// NewRouter initializes application routes using the provided handlers.
func NewRouter(routeFuncs []RouteRegisterFunc, handlers []handler.Handler) *Router {
	mux := http.NewServeMux()

	// Register each route module dynamically
	for i, registerFunc := range routeFuncs {
		registerFunc(mux, &handlers[i])
	}

	// Apply middleware to limit request body size (1MB limit)
	limitedMux := middleware.LimitRequestBodyMiddleware(1 << 20)(mux)

	return &Router{
		Mux:            mux,
		LimitedHandler: limitedMux,
	}
}
