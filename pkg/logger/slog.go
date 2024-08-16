package logger

import "log/slog"

// std is a simple implementation of Logger for log std library.
type slo struct {
	slo *slog.Logger
}

var _ Logger = &slo{slo: slog.Default()} // ensure interface is implemented

// Slog returns the default slog logger (log/slog library).
func Slog() Logger {
	return &slo{slog.Default()}
}

// Info logs with slog logger using Info function.
func (s *slo) Info(args ...any) {
	if len(args) == 0 {
		return
	}

	msg, ok := args[0].(string)
	if !ok {
		s.slo.Info("", args...)
		return
	}
	s.slo.Info(msg, args[1:]...)
}

// Infof logs with slog logger using Info function.
func (s *slo) Infof(msg string, args ...any) {
	s.slo.Info(msg, args...)
}

// Warn logs with slog logger using Warn function.
func (s *slo) Warn(args ...any) {
	if len(args) == 0 {
		return
	}

	msg, ok := args[0].(string)
	if !ok {
		s.slo.Warn("", args...)
		return
	}
	s.slo.Warn(msg, args[1:]...)
}

// Warnf logs with slog logger using Warn function.
func (s *slo) Warnf(msg string, args ...any) {
	s.slo.Warn(msg, args...)
}
