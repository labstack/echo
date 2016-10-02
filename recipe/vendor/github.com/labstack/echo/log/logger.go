package log

import (
	"io"

	"github.com/labstack/gommon/log"
)

type (
	// Logger defines the logging interface.
	Logger interface {
		SetOutput(io.Writer)
		SetLevel(log.Lvl)
		Print(...interface{})
		Printf(string, ...interface{})
		Printj(log.JSON)
		Debug(...interface{})
		Debugf(string, ...interface{})
		Debugj(log.JSON)
		Info(...interface{})
		Infof(string, ...interface{})
		Infoj(log.JSON)
		Warn(...interface{})
		Warnf(string, ...interface{})
		Warnj(log.JSON)
		Error(...interface{})
		Errorf(string, ...interface{})
		Errorj(log.JSON)
		Fatal(...interface{})
		Fatalj(log.JSON)
		Fatalf(string, ...interface{})
	}
)
