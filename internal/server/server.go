// Package server listens on the port for HTTP 1/2 connections and sends them to the router.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/henryhall897/golang-todo-app/internal/config"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// HTTPServer is a server that listens for HTTP/1.x and also supports HTTP/2 cleartext (h2c).
type HTTPServer struct {
	cfg *config.ServerConfig
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(config *config.ServerConfig) *HTTPServer {
	return &HTTPServer{cfg: config}
}

// Serve starts the HTTP server and supports HTTP/2 cleartext (h2c).
func (s *HTTPServer) Serve(ctx context.Context, handler http.Handler) error {
	s.cfg.Logger.Infof("Starting server on %s:%s", s.cfg.BindAddress, s.cfg.Port)

	// Create an HTTP/2 server
	h2s := &http2.Server{
		MaxConcurrentStreams: 250,
	}

	// Wrap the handler to support h2c (HTTP/2 without TLS)
	wrappedHandler := h2c.NewHandler(handler, h2s)

	// Initialize HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.cfg.BindAddress, s.cfg.Port),
		Handler: wrappedHandler,
	}

	// Channel to capture errors
	errCh := make(chan error, 1)

	// Goroutine to handle graceful shutdown
	go func() {
		<-ctx.Done()

		s.cfg.Logger.Info("Shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	// Start the HTTP server
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	s.cfg.Logger.Info("Server stopped")

	// Return any shutdown errors
	select {
	case err := <-errCh:
		return fmt.Errorf("failed to shutdown: %w", err)
	default:
		return nil
	}
}
