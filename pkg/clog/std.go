package clog

import "log"

// stdlog is a simple implementation of Logger for log stdlog library.
type stdlog struct {
	log *log.Logger
}

var _log Logger = &stdlog{log.Default()} // ensure interface is implemented

// Std returns the default std logger (log library).
func Std() Logger {
	return _log
}

// StdWith returns the Logger interface with input std logger.
func StdWith(logger *log.Logger) Logger {
	return &stdlog{logger}
}

// Debugf logs with std logger using Printf function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *stdlog) Debugf(msg string, args ...any) {
	s.print(msg, args...)
}

// Errorf logs with std logger using Printf function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *stdlog) Errorf(msg string, args ...any) {
	s.print(msg, args...)
}

// Infof logs with std logger using Printf function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *stdlog) Infof(msg string, args ...any) {
	s.print(msg, args...)
}

// Warnf logs with std logger using Printf function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *stdlog) Warnf(msg string, args ...any) {
	s.print(msg, args...)
}

func (s *stdlog) print(msg string, args ...any) {
	s.log.Printf(msg+"\n", args...)
}
