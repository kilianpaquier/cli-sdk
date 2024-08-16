package logger

// Logger is a simplified interface for logging purposes in craft features.
//
// It's a interface because it would allow anyone to use any logger implementation (logrus, log, slog, etc.).
//
// In case you don't need or want a specific implementation,
// you can use default implementations Slog or Std which respectively use log/slog and log builtin libraries.
type Logger interface {
	// Info should log with the INFO level.
	Info(...any)

	// Infof should log with the INFO level and use format subtitution to take care of input args.
	Infof(string, ...any)

	// Warn should log with the WARN level.
	Warn(...any)

	// Warnf should log with the WARN level and use format subtitution to take care of input args.
	Warnf(string, ...any)
}
