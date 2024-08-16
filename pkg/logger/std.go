package logger

import "log"

// std is a simple implementation of Logger for log std library.
type std struct {
	std *log.Logger
}

var _ Logger = &std{std: log.Default()} // ensure interface is implemented

// Std returns the default std logger (log library).
func Std() Logger {
	return &std{log.Default()}
}

// Info logs with std logger using Println function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *std) Info(args ...any) {
	s.std.Println(args...)
}

// Infof logs with std logger using Printf function
// with newline is automatically added to the end of msg.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *std) Infof(msg string, args ...any) {
	s.std.Printf(msg+"\n", args...)
}

// Warn logs with std logger using Println function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *std) Warn(args ...any) {
	s.Info(args...)
}

// Warnf logs with std logger using Printf function.
//
// No logging level is involved since base std library doesn't handle logging level.
func (s *std) Warnf(msg string, args ...any) {
	s.Infof(msg, args...)
}