package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"golang-todo-app/internal/core/logging"
	"golang-todo-app/internal/routes"
	"golang-todo-app/internal/users"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables
	dbURL := os.Getenv("DATABASE_URL")
	serverBindAddress := os.Getenv("SERVER_BIND_ADDRESS")
	serverPort := os.Getenv("SERVER_PORT")
	//corsOrigin := os.Getenv("CORS_ORIGIN")
	logLevel := os.Getenv("LOG_LEVEL")

	// Initialize logger
	logger, err := initializeLogger(logLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	sugarLogger := logger.Sugar()

	zapLogger := logging.ZapLogger{SugaredLogger: sugarLogger}

	// Apply database migrations
	if err := applyMigrations(dbURL); err != nil {
		sugarLogger.Fatalf("Failed to apply migrations: %v", err)
	}

	// Initialize database connection pool
	pool, err := initializeDatabasePool(dbURL)
	if err != nil {
		sugarLogger.Fatalf("Failed to connect to the database: %v", err)
	}
	defer pool.Close()

	// Initialize stores
	userStore := users.New(pool)

	// Setup routes and handlers
	router := mux.NewRouter()
	routes.RegisterUserRoutes(router, userStore, &zapLogger)

	// Apply CORS middleware (optional)
	// corsMiddleware := middleware.CORS(corsOrigin)
	// router.Use(corsMiddleware)

	// Start the server with graceful shutdown
	srv := &http.Server{
		Addr:         serverBindAddress + ":" + serverPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		sugarLogger.Infof("Server running on http://%s:%s", serverBindAddress, serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugarLogger.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	sugarLogger.Info("Shutting down server...")

	if err := srv.Shutdown(ctx); err != nil {
		sugarLogger.Fatalf("Server shutdown failed: %v", err)
	}
	sugarLogger.Info("Server shut down gracefully.")
}

// initializeLogger sets up the logger based on the log level from environment variables
func initializeLogger(logLevel string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	switch logLevel {
	case "DEBUG":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "INFO":
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "ERROR":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	return cfg.Build()
}

// applyMigrations applies all up migrations from the migrations directory
func applyMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://migrations", // Path to your migration files
		databaseURL,
	)
	if err != nil {
		return err
	}

	// Run the migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Database migrations applied successfully.")
	return nil
}

// initializeDatabasePool sets up the database connection pool
func initializeDatabasePool(dbURL string) (*pgxpool.Pool, error) {
	maxConn, _ := strconv.Atoi(os.Getenv("POOL_MAX_CONN"))
	minConn, _ := strconv.Atoi(os.Getenv("POOL_MIN_CONN"))

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	// Apply pool configuration from environment variables
	config.MaxConns = int32(maxConn)
	config.MinConns = int32(minConn)

	return pgxpool.NewWithConfig(context.Background(), config)
}
