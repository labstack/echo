package log

import "io"

type (
	// Logger defines the logging interface.
	Logger interface {
		SetOutput(io.Writer)
		SetLevel(uint8)
		Print(...interface{})
		Printf(string, ...interface{})
		Debug(...interface{})
		Debugf(string, ...interface{})
		Info(...interface{})
		Infof(string, ...interface{})
		Warn(...interface{})
		Warnf(string, ...interface{})
		Error(...interface{})
		Errorf(string, ...interface{})
		Fatal(...interface{})
		Fatalf(string, ...interface{})
	}
)
