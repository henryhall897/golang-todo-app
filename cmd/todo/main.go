// Package main is the entry point to the application
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang-todo-app/internal/core/logging"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	envconfig "github.com/sethvargo/go-envconfig"
)

var (
	version   string
	buildDate string
)

// TODO
// Config project configuration
type Config struct {
	// Database dbpool.Config
	// Server   server.Config
}

// TODO
// Log prints the configuration to the log
func (cfg *Config) Log(ctx context.Context) {
	logger := logging.GetLogger(ctx)

	logger.Infow("Configuration")
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

	logger := logging.InitializeLogger("debug", false).Sugar()

	defer func() {
		done()
		if r := recover(); r != nil {
			logger.Errorw("application panic", "panic", r)
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

func run(ctx context.Context, logger *zap.SugaredLogger) error {

	logger.Debugw("Read in Configuration",
		"version", version,
		"buildDate", buildDate)
	cfg := Config{}
	if err := Setup(ctx, &cfg); err != nil {
		return fmt.Errorf("failed to read the configuration from the environment: %w", err)
	}
	logger.Infow("go-crud-example", "config", cfg)

	// logger.Debug("Open Database")
	// pool, err := dbpool.New(ctx, logger, &cfg.Database)
	// if err != nil {
	// 	return fmt.Errorf("unable to open the database: %w", err)
	// }

	// h := handler.BuildHandler(ctx, logger, models.NewDBStore(pool))
	// srv := server.NewHTTPServer(logger, &cfg.Server)
	// return srv.Serve(ctx, h)
	return nil
}

func Setup(ctx context.Context, config any) error {
	return envconfig.Process(ctx, config)
}
