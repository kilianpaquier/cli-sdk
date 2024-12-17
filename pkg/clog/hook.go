package clog

// Level is a type for logging levels.
//
// It's really only used with Hook function type.
type Level string

const (
	// LevelDebug is the DEBUG logging level.
	LevelDebug Level = "debug"

	// LevelError is the ERROR logging level.
	LevelError Level = "error"

	// LevelInfo is the INFO logging level.
	LevelInfo Level = "info"

	// LevelWarn is the WARN logging level.
	LevelWarn Level = "warn"
)

type Record struct {
	Args  []any
	Level Level
	Msg   string
}

// Hook is a function that can be used to modify the message before it is logged.
type Hook func(level Level, msg string) string

// hookLogger is a Logger decorator that wraps another Logger
// and executes a Hook before logging.
type hookLogger struct {
	logger Logger
	hook   Hook
}

var _ Logger = &hookLogger{} // ensure interface is implemented

// NewHook returns a Logger that wraps the input logger
// and applies the input hook to the message before logging it.
func NewHook(logger Logger, hook Hook) Logger {
	return &hookLogger{logger: logger, hook: hook}
}

// Debugf executes the hook before logging the message with wrapped logger.
func (h *hookLogger) Debugf(format string, args ...any) {
	h.logger.Debugf(h.hook(LevelDebug, format), args...)
}

// Errorf executes the hook before logging the message with wrapped logger.
func (h *hookLogger) Errorf(format string, args ...any) {
	h.logger.Errorf(h.hook(LevelError, format), args...)
}

// Infof executes the hook before logging the message with wrapped logger.
func (h *hookLogger) Infof(format string, args ...any) {
	h.logger.Infof(h.hook(LevelInfo, format), args...)
}

// Warnf executes the hook before logging the message with wrapped logger.
func (h *hookLogger) Warnf(format string, args ...any) {
	h.logger.Warnf(h.hook(LevelWarn, format), args...)
}

const (
	cyan   = "\033[36m"
	gray   = "\033[90m"
	red    = "\033[31m"
	reset  = "\033[0m"
	yellow = "\033[33m"
)

// ColorHook is a Hook that colors the message based on the logging level.
func ColorHook(level Level, msg string) string {
	switch level {
	case LevelDebug:
		return gray + msg + reset
	case LevelError:
		return red + msg + reset
	case LevelInfo:
		return cyan + msg + reset
	case LevelWarn:
		return yellow + msg + reset
	default:
		return msg
	}
}

var _ Hook = ColorHook // ensure interface is implemented
