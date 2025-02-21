package main

import (
	"context"
	"log"

	"github.com/henryhall897/golang-todo-app/internal/config"
	"github.com/henryhall897/golang-todo-app/internal/core/logging"
	"github.com/henryhall897/golang-todo-app/internal/database"
	"github.com/henryhall897/golang-todo-app/internal/server"
	"github.com/henryhall897/golang-todo-app/internal/users"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	ctx := context.Background()
	// Load environment variables
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	// Initialize logger
	logger := logging.InitializeLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	sugarLogger := logger.Sugar()

	// Apply database migrations
	if err := database.ApplyMigrations(cfg.DatabaseURL); err != nil {
		sugarLogger.Fatalf("Failed to apply migrations: %v", err)
	}

	// Initialize database connection pool
	pool, err := database.InitializeDatabasePool(sugarLogger)
	if err != nil {
		sugarLogger.Fatalf("Failed to connect to the database: %v", err)
	}
	defer pool.Close()

	// Initialize stores
	userStore := users.New(pool)

	// Initialize server
	srv := server.NewServer(sugarLogger, cfg.ServerBindAddress, cfg.ServerPort, userStore)
	srv.Start()
}
