// Package main is the entry point to the application
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang-todo-app/pkg/logging"

	"github.com/joho/godotenv"

	envconfig "github.com/sethvargo/go-envconfig"
)

var (
	version   string
	buildDate string
)

// Config project configuration
type Config struct {
	// Database dbpool.Config
	// Server   server.Config
}

// Log prints the configuration to the log
func (cfg *Config) Log(ctx context.Context) {
	logger := logging.GetLogger(ctx)

	logger.Infow(ctx, "Configuration")
	// "database", cfg.Database,
	// "server", cfg.Server)
}

func main() {
	err := godotenv.Overload("../../.env")
	if err != nil {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(dir)
		}
	}

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	logger := logging.InitializeLogger("debug", false)

	defer func() {
		done()
		if r := recover(); r != nil {
			logger.ErrorContext(ctx, "application panic", "panic", r)
			os.Exit(1)
		}
	}()

	err = run(ctx, logger)
	done()

	if err != nil {
		logger.Error("run result", "err", err)
		os.Exit(1)
	}
	logger.Info("successful shutdown")
}

func run(ctx context.Context, logger *slog.Logger) error {

	logger.Debug("Read in Configuration",
		"version", version,
		"buildDate", buildDate)
	cfg := Config{}
	if err := Setup(ctx, &cfg); err != nil {
		return fmt.Errorf("failed to read the configuration from the environment: %w", err)
	}
	logger.Info("go-crud-example", "config", cfg)

	// logger.Debug("Open Database")
	// pool, err := dbpool.New(ctx, logger, &cfg.Database)
	// if err != nil {
	// 	return fmt.Errorf("unable to open the database: %w", err)
	// }

	// h := handler.BuildHandler(ctx, logger, models.NewDBStore(pool))
	// srv := server.NewHTTPServer(logger, &cfg.Server)
	// return srv.Serve(ctx, h)
}

func Setup(ctx context.Context, config interface{}) error {
	return envconfig.ProcessWith(ctx, config, envconfig.OsLookuper())
}
