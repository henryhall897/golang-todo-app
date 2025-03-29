package config

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
)

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
	MinConns    int32  `env:"POSTGRES_POOL_MIN_CONN,default=1"`
	MaxConns    int32  `env:"POSTGRES_POOL_MAX_CONN,default=10"`
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	BindAddress string `env:"BIND_ADDRESS,default=0.0.0.0"`
	Port        string `env:"PORT,default=8080"`
	Logger      *zap.SugaredLogger
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string `env:"LOG_LEVEL,default=info"`
	Format bool   `env:"LOG_FORMAT,default=false"`
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Address  string `env:"REDIS_ADDRESS,required"`
	Password string `env:"REDIS_PASSWORD,default="`
	DB       int    `env:"REDIS_DB,default=0"`
}

// AppConfig holds the complete application configuration
type AppConfig struct {
	Database DatabaseConfig
	Server   ServerConfig
	Logger   LoggingConfig
	Redis    RedisConfig
}

// LoadConfig loads the entire configuration from environment variables
func LoadConfig(ctx context.Context) (*AppConfig, error) {
	var cfg AppConfig

	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}
