package clog

import "log/slog"

// std is a simple implementation of Logger for log std library.
type stdslog struct {
	log *slog.Logger
}

var _slog Logger = &stdslog{slog.Default()} // ensure interface is implemented

// Slog returns the default slog logger (slog library).
func Slog() Logger {
	return _slog
}

// SlogWith returns the Logger interface with input slog logger.
func SlogWith(logger *slog.Logger) Logger {
	return &stdslog{logger}
}

// Debugf logs with slog logger using Debug function.
func (s *stdslog) Debugf(msg string, args ...any) {
	s.log.Debug(msg, args...)
}

// Errorf logs with slog logger using Debug function.
func (s *stdslog) Errorf(msg string, args ...any) {
	s.log.Error(msg, args...)
}

// Infof logs with slog logger using Debug function.
func (s *stdslog) Infof(msg string, args ...any) {
	s.log.Info(msg, args...)
}

// Warnf logs with slog logger using Debug function.
func (s *stdslog) Warnf(msg string, args ...any) {
	s.log.Warn(msg, args...)
}
