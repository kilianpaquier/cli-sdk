package clog

import (
	"fmt"
	"log/slog"
)

// std is a simple implementation of Logger for log std library.
type stdslog struct {
	logger *slog.Logger
}

// Slog returns the Logger interface with input slog logger.
func Slog(logger *slog.Logger) Logger {
	return &stdslog{logger}
}

// Debugf logs with slog logger using Debug function.
func (s *stdslog) Debugf(format string, args ...any) {
	s.logger.Debug(fmt.Sprintf(format, args...))
}

// Errorf logs with slog logger using Debug function.
func (s *stdslog) Errorf(format string, args ...any) {
	s.logger.Error(fmt.Sprintf(format, args...))
}

// Infof logs with slog logger using Debug function.
func (s *stdslog) Infof(format string, args ...any) {
	s.logger.Info(fmt.Sprintf(format, args...))
}

// Warnf logs with slog logger using Debug function.
func (s *stdslog) Warnf(format string, args ...any) {
	s.logger.Warn(fmt.Sprintf(format, args...))
}
