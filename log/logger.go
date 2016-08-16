package log

import (
	"io"

	"golang.org/x/net/context"
)

type (
	// Logger defines the logging interface.
	Logger interface {
		SetOutput(io.Writer)
		Output() io.Writer
		SetLevel(Lvl)
		Level() Lvl
		Print(...interface{})
		Printf(string, ...interface{})
		Printj(JSON)
		Debug(...interface{})
		Debugf(string, ...interface{})
		Debugj(JSON)
		Info(...interface{})
		Infof(string, ...interface{})
		Infoj(JSON)
		Warn(...interface{})
		Warnf(string, ...interface{})
		Warnj(JSON)
		Error(...interface{})
		Errorf(string, ...interface{})
		Errorj(JSON)
		Fatal(...interface{})
		Fatalj(JSON)
		Fatalf(string, ...interface{})
	}

	Lvl uint8

	JSON map[string]interface{}

	Option func(*options)

	options struct {
		prefix string
		ctx    context.Context
	}
)

func Prefix(prefix string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

const (
	DEBUG Lvl = iota
	INFO
	WARN
	ERROR
	FATAL
	OFF
)
