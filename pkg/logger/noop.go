package logger

type noop struct{}

var nooplog Logger = &noop{} // ensure interface is implemented

// Noop returns an empty logger which will do nothing.
func Noop() Logger {
	return nooplog
}

// Infof does nothing.
func (*noop) Infof(string, ...any) {}

// Warnf does nothing.
func (*noop) Warnf(string, ...any) {}
