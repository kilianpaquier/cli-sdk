package clog

import "log"

// std is a simple implementation of Logger for log std library.
type std struct {
	l *log.Logger
}

var _log Logger = &std{log.Default()} // ensure interface is implemented

// Std returns the default std logger (log library).
func Std() Logger {
	return _log
}

// StdWith returns the Logger interface with input std logger.
func StdWith(l *log.Logger) Logger {
	return &std{l}
}

// Infof logs with std logger using Printf function
// with newline is automatically added to the end of msg.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *std) Infof(msg string, args ...any) {
	s.l.Printf(msg+"\n", args...)
}

// Warnf logs with std logger using Printf function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *std) Warnf(msg string, args ...any) {
	s.Infof(msg, args...)
}
