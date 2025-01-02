package dbpool

import (
	"context"

	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

type zapAdapter struct {
	logger *zap.Logger
}

func (a *zapAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	fields := make([]zap.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	switch level {
	case tracelog.LogLevelTrace:
		a.logger.Debug(msg, fields...)
	case tracelog.LogLevelDebug:
		a.logger.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		a.logger.Info(msg, fields...)
	case tracelog.LogLevelWarn:
		a.logger.Warn(msg, fields...)
	case tracelog.LogLevelError:
		a.logger.Error(msg, fields...)
	default:
		a.logger.Error(msg, fields...)
	}
}
