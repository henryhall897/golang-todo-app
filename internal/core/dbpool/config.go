package dbpool

import (
	"fmt"

	"go.uber.org/zap"
)

// Config holds the database configuration.
type Config struct {
	Logging      bool   `env:"DB_LOGGING_ENABLE, default=false"`
	Host         string `env:"POSTGRES_HOST, default=localhost" json:"host"`
	Port         string `env:"POSTGRES_PORT, default=5432" json:"port"`
	User         string `env:"POSTGRES_USER" json:"user"`
	Password     string `env:"POSTGRES_PASSWORD" json:"-"` // Hide from logging
	DatabaseName string `env:"POSTGRES_DB" json:"database_name"`
	MaxConns     int32  `env:"POOL_MAX_CONN, default=2"`
	MinConns     int32  `env:"POOL_MIN_CONN, default=1"`
}

// ConnectString constructs the connection string from ENV.
// Example: postgres://username:password@localhost:5432/database_name
func (c Config) ConnectString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.User, c.Password, c.Host, c.Port, c.DatabaseName)
}

// LogConfig logs the database configuration using Uber Zap.
func (c Config) LogConfig(logger *zap.SugaredLogger) {
	logger.Info("Database Configuration",
		zap.Bool("logging", c.Logging),
		zap.String("host", c.Host),
		zap.String("port", c.Port),
		zap.String("user", "********"),     // Hide sensitive information
		zap.String("password", "********"), // Hide sensitive information
		zap.String("database_name", c.DatabaseName),
		zap.Int32("max_conns", c.MaxConns),
		zap.Int32("min_conns", c.MinConns),
	)
}

// ConfigToZapFields converts the configuration into Zap fields for structured logging.
func (c Config) ConfigToZapFields() []zap.Field {
	return []zap.Field{
		zap.Bool("logging", c.Logging),
		zap.String("host", c.Host),
		zap.String("port", c.Port),
		zap.String("user", "********"),     // Hide sensitive information
		zap.String("password", "********"), // Hide sensitive information
		zap.String("database_name", c.DatabaseName),
		zap.Int32("max_conns", c.MaxConns),
		zap.Int32("min_conns", c.MinConns),
	}
}
