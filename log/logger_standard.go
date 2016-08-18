// +build !appengine

package log

import (
	"io"

	glog "github.com/labstack/gommon/log"
)

type stdLogger struct {
	l *glog.Logger
}

func (l *stdLogger) SetOutput(w io.Writer) {
	l.l.SetOutput(w)
}

func (l *stdLogger) Output() io.Writer {
	return l.l.Output()
}

func (l *stdLogger) SetLevel(v Lvl) {
	l.l.SetLevel(glog.Lvl(v))
}

func (l *stdLogger) Level() Lvl {
	return Lvl(l.l.Level())
}

func (l *stdLogger) Print(i ...interface{}) {
	l.l.Print(i...)
}

func (l *stdLogger) Printf(format string, args ...interface{}) {
	l.l.Printf(format, args...)
}

func (l *stdLogger) Printj(j JSON) {
	l.l.Printj(glog.JSON(j))
}

func (l *stdLogger) Debug(i ...interface{}) {
	l.l.Debug(i...)
}

func (l *stdLogger) Debugf(format string, args ...interface{}) {
	l.l.Debugf(format, args...)
}

func (l *stdLogger) Debugj(j JSON) {
	l.l.Debugj(glog.JSON(j))
}

func (l *stdLogger) Info(i ...interface{}) {
	l.l.Info(i...)
}

func (l *stdLogger) Infof(format string, args ...interface{}) {
	l.l.Infof(format, args...)
}

func (l *stdLogger) Infoj(j JSON) {
	l.l.Infoj(glog.JSON(j))
}

func (l *stdLogger) Warn(i ...interface{}) {
	l.l.Warn(i...)
}

func (l *stdLogger) Warnf(format string, args ...interface{}) {
	l.l.Warnf(format, args...)
}

func (l *stdLogger) Warnj(j JSON) {
	l.l.Warnj(glog.JSON(j))
}

func (l *stdLogger) Error(i ...interface{}) {
	l.l.Error(i...)
}

func (l *stdLogger) Errorf(format string, args ...interface{}) {
	l.l.Errorf(format, args...)
}

func (l *stdLogger) Errorj(j JSON) {
	l.l.Errorj(glog.JSON(j))
}

func (l *stdLogger) Fatal(i ...interface{}) {
	l.l.Fatal(i...)
}

func (l *stdLogger) Fatalf(format string, args ...interface{}) {
	l.l.Fatalf(format, args...)
}

func (l *stdLogger) Fatalj(j JSON) {
	l.l.Fatalj(glog.JSON(j))
}

func New(opts ...Option) Logger {
	opt := &options{prefix: "echo"}
	for _, o := range opts {
		o(opt)
	}
	return &stdLogger{glog.New(opt.prefix)}
}
