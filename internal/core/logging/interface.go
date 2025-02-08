package logging

import "go.uber.org/zap"

// ZapLogger wraps a SugaredLogger to implement the Logger interface.
type ZapLogger struct {
	SugaredLogger *zap.SugaredLogger
}

func (z *ZapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Errorw(msg, keysAndValues...)
}

func (z *ZapLogger) Infow(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Infow(msg, keysAndValues...)
}

func (z *ZapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	z.SugaredLogger.Debugw(msg, keysAndValues...)
}

// Implement other methods as needed

// Logger defines the logging mechanism interface.
type Logger interface {
	Errorw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	// Add other methods as needed
}
