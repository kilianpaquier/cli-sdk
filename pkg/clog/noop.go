package clog

type noop struct{}

var _noop Logger = &noop{} // ensure interface is implemented

// Noop returns an empty logger which will do nothing.
func Noop() Logger {
	return _noop
}

// Debugf does nothing.
func (*noop) Debugf(string, ...any) {}

// Errorf does nothing.
func (*noop) Errorf(string, ...any) {}

// Infof does nothing.
func (*noop) Infof(string, ...any) {}

// Warnf does nothing.
func (*noop) Warnf(string, ...any) {}
