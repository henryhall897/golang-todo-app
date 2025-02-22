package server

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/henryhall897/golang-todo-app/internal/middleware"
	"github.com/henryhall897/golang-todo-app/internal/users"
	"go.uber.org/zap"
)

// Server encapsulates all server dependencies and configuration.
type Server struct {
	Logger      *zap.SugaredLogger
	BindAddress string
	Port        string
	UserStore   users.UserRepository
	httpServer  *http.Server
}

// NewServer initializes a new Server instance.
func NewServer(logger *zap.SugaredLogger, bindAddress string, port string, userStore users.UserRepository) *Server {
	return &Server{
		Logger:      logger,
		BindAddress: bindAddress,
		Port:        port,
		UserStore:   userStore,
	}
}

// Start initializes routes and starts the HTTP server with graceful shutdown.
func (s *Server) Start() {
	mux := http.NewServeMux()

	// Register user routes
	users.RegisterUserRoutes(mux, s.UserStore, s.Logger)

	// Apply middleware to limit request body size (1MB limit)
	limitedMux := middleware.LimitRequestBodyMiddleware(1 << 20)(mux)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         s.BindAddress + ":" + s.Port,
		Handler:      limitedMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run server in a separate goroutine
	go func() {
		s.Logger.Infof("Server running on http://%s:%s", s.BindAddress, s.Port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for termination signal
	<-ctx.Done()
	s.Shutdown(ctx)
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) {
	s.Logger.Info("Shutting down server...")

	// Attempt graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.Logger.Fatalf("Server shutdown failed: %v", err)
	}
	s.Logger.Info("Server shut down gracefully.")
}
