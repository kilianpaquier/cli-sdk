package logger

type noop struct{}

var _ Logger = &noop{} // ensure interface is implemented

// Noop returns an empty logger which will do nothing.
func Noop() Logger {
	return &noop{}
}

// Infof does nothing.
func (*noop) Infof(string, ...any) {}

// Warnf does nothing.
func (*noop) Warnf(string, ...any) {}
