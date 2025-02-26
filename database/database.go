package database

import (
	"context"
	"fmt"

	"github.com/henryhall897/golang-todo-app/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
)

// InitializeDatabasePool sets up the PostgreSQL connection pool
func InitializeDatabasePool(logger *zap.SugaredLogger) (*pgxpool.Pool, error) {
	var cfg config.DatabaseConfig

	// Load environment variables
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		logger.Errorw("Failed to load environment variables", "error", err)
		return nil, fmt.Errorf("failed to load environment variables: %v", err)
	}

	// Parse database config
	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Errorw("Failed to parse database config", "error", err)
		return nil, fmt.Errorf("failed to parse database config: %v", err)
	}

	// Apply pool configuration
	dbConfig.MinConns = cfg.MinConns
	dbConfig.MaxConns = cfg.MaxConns

	// Initialize the database pool
	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		logger.Errorw("Failed to create database pool", "error", err)
		return nil, fmt.Errorf("failed to create database pool: %v", err)
	}

	logger.Infow("Database pool successfully initialized", "minConns", cfg.MinConns, "maxConns", cfg.MaxConns)
	return pool, nil
}
