package router

import "net/http"

// Handlers is an interface that defines the methods for handling routes.
type Routes interface {
	RegisterRoutes(router *http.ServeMux)
}
