package router

import (
	"net/http"

	"github.com/henryhall897/golang-todo-app/internal/middleware"
)

// router manages the routes for the application.
type Router struct {
	Mux            *http.ServeMux
	LimitedHandler http.Handler
}

// NewRouter initializes application routes using the provided handlers.
func NewRouter(routes []Routes) *Router {
	mux := http.NewServeMux()

	// Register each route module dynamically
	for _, route := range routes {
		route.RegisterRoutes(mux)
	}

	// Apply middleware to limit request body size (1MB limit)
	limitedMux := middleware.LimitRequestBodyMiddleware(1 << 20)(mux)

	return &Router{
		Mux:            mux,
		LimitedHandler: limitedMux,
	}
}
