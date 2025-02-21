package config

import (
	"context"
	"log"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	DatabaseURL       string `env:"DATABASE_URL,required"`
	MinConns          int32  `env:"POSTGRES_POOL_MIN_CONN,default=1"`
	MaxConns          int32  `env:"POSTGRES_POOL_MAX_CONN,default=10"`
	ServerBindAddress string `env:"SERVER_BIND_ADDRESS,default=0.0.0.0"`
	ServerPort        string `env:"SERVER_PORT,default=8080"`
	LogLevel          string `env:"LOG_LEVEL,default=info"`
	LogFormat         bool   `env:"LOG_FORMAT,default=false"`
}

// LoadConfig initializes the Config struct by reading from environment variables
func LoadConfig(ctx context.Context) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
		return nil, err
	}
	return &cfg, nil
}
