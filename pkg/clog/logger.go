package clog

import "context"

// Logger is a simplified interface for logging purposes.
type Logger interface {
	// Debugf logs with the DEBUG level.
	Debugf(format string, args ...any)

	// Errorf logs with the ERROR level.
	Errorf(format string, args ...any)

	// Infof logs with the INFO level.
	Infof(format string, args ...any)

	// Warnf logs with the WARN level.
	Warnf(format string, args ...any)
}

type loggerKeyType string

// LoggerKey is the context key for the logger.
const LoggerKey loggerKeyType = "logger"

// GetLogger returns the context logger.
//
// By default it will be clog.Noop, but it can be set with WithLogger.
func GetLogger(ctx context.Context) Logger {
	log, ok := ctx.Value(LoggerKey).(Logger)
	if !ok {
		return Noop()
	}
	return log
}
