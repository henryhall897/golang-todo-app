package logging

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const loggerKey contextKey = "logger"

var defaultLogger *zap.Logger

// InitializeLogger creates a new default Zap logger with the specified level and format.
// Parameters:
// - level: Log level (e.g., "debug", "info", "warn", "error").
// - jsonFormat: If true, log output will be in JSON format.
func InitializeLogger(level string, jsonFormat bool) *zap.Logger {
	var zapConfig zap.Config

	if jsonFormat {
		// Production config with JSON formatting.
		zapConfig = zap.NewProductionConfig()
	} else {
		// Development config with console formatting.
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level dynamically.
	switch level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	var err error
	defaultLogger, err = zapConfig.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	return defaultLogger
}

// GetDefaultLogger returns the default logger.
func GetDefaultLogger() *zap.Logger {
	if defaultLogger == nil {
		InitializeLogger("info", false) // Fallback initialization in case it's not explicitly called.
	}
	return defaultLogger
}

// WithLogger sets the given logger into the context.
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLogger retrieves the logger from the context. If no logger is found, it returns the default logger.
func GetLogger(ctx context.Context) *zap.SugaredLogger {
	logger, ok := ctx.Value(loggerKey).(*zap.SugaredLogger)
	if !ok {
		return GetDefaultLogger().Sugar()
	}
	return logger
}
