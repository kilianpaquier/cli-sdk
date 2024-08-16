package clog

// Logger is a simplified interface for logging purposes.
type Logger interface {
	// Debugf logs with the DEBUG level
	// and uses format subtitution to take care of input args.
	Debugf(string, ...any)

	// Errorf logs with the ERROR level
	// and uses format subtitution to take care of input args.
	Errorf(string, ...any)

	// Infof logs with the INFO level
	// and uses format subtitution to take care of input args.
	Infof(string, ...any)

	// Warnf logs with the WARN level
	// and uses format subtitution to take care of input args.
	Warnf(string, ...any)
}
