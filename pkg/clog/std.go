package clog

import "log"

// stdlog is a simple implementation of Logger for log stdlog library.
type stdlog struct {
	logger *log.Logger
}

// Std returns the Logger interface with input std logger.
func Std(logger *log.Logger) Logger {
	return &stdlog{logger: logger}
}

// Debugf logs with std logger using Printf function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *stdlog) Debugf(format string, args ...any) {
	s.print(format, args...)
}

// Errorf logs with std logger using Printf function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *stdlog) Errorf(format string, args ...any) {
	s.print(format, args...)
}

// Infof logs with std logger using Printf function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *stdlog) Infof(format string, args ...any) {
	s.print(format, args...)
}

// Warnf logs with std logger using Printf function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *stdlog) Warnf(format string, args ...any) {
	s.print(format, args...)
}

func (s *stdlog) print(format string, args ...any) {
	s.logger.Printf(format+"\n", args...)
}
