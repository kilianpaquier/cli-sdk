package clog

// Logger is a simplified interface for logging purposes.
//
// It's a interface because it would allow anyone to use any logger implementation (logrus, log, slog, etc.).
type Logger interface {
	// Infof logs with the INFO level and use format subtitution to take care of input args.
	Infof(string, ...any)

	// Warnf logs with the WARN level and use format subtitution to take care of input args.
	Warnf(string, ...any)
}
