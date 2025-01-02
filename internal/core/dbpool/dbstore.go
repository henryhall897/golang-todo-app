package dbpool

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

// New creates a new pgx pool connection from the config
func New(ctx context.Context, logger *zap.Logger, cfg *Config) (*pgxpool.Pool, error) {
	// Parse the connection string into pgxpool config
	cConfig, err := pgxpool.ParseConfig(cfg.ConnectString())
	if err != nil {
		return nil, fmt.Errorf("unable to parse pgx connection: %w", err)
	}

	// Set connection pool configurations
	cConfig.MaxConns = cfg.MaxConns
	cConfig.MinConns = cfg.MinConns

	// Enable logging if configured
	if cfg.Logging {
		cConfig.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   &zapAdapter{logger: logger},
			LogLevel: tracelog.LogLevelDebug, // Adjust the log level as needed
		}
	}

	// Establish the connection pool
	dbpool, err := pgxpool.NewWithConfig(ctx, cConfig)
	if err != nil {
		logger.Error("Failed to connect to the database", zap.Error(err))
		return nil, err
	}

	logger.Debug("Database connection pool created successfully")
	return dbpool, nil
}
