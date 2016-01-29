package echo

import (
	"fmt"
	"github.com/labstack/gommon/log"
)

type Logger interface {
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

type GommonLogger struct {
	logger *log.Logger
}

func (l *GommonLogger) Debug(values ...interface{}) {
	l.logger.Debug(fmt.Sprint(values...))
}

func (l *GommonLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(format, args...)
}

func (l *GommonLogger) Info(values ...interface{}) {
	l.logger.Info(fmt.Sprint(values...))
}

func (l *GommonLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(format, args...)
}

func (l *GommonLogger) Warn(values ...interface{}) {
	l.logger.Warn(fmt.Sprint(values...))
}

func (l *GommonLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn(format, args...)
}

func (l *GommonLogger) Error(values ...interface{}) {
	l.logger.Error(fmt.Sprint(values...))
}

func (l *GommonLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(format, args...)
}

func (l *GommonLogger) Fatal(values ...interface{}) {
	l.logger.Fatal(fmt.Sprint(values...))
}

func (l *GommonLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal(format, args...)
}

func DefaultLogger() Logger {
	return &GommonLogger{logger: log.New("echo")}
}
