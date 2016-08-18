// +build appengine

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"
)

type gaeLogger struct {
	ctx context.Context
	v   Lvl
	b   *bytes.Buffer
}

func (l *gaeLogger) SetContext(ctx context.Context) {
	l.ctx = ctx
}

func (l *gaeLogger) json(j JSON) (res string) {
	json.NewEncoder(l.b).Encode(j)
	res = l.b.String()
	l.b.Reset()
	return
}

func (l *gaeLogger) SetOutput(w io.Writer) {
	panic("echo: Unsupport SetOutput")
}

func (l *gaeLogger) Output() io.Writer {
	panic("echo: Unsupport Output")
}

func (l *gaeLogger) SetLevel(v Lvl) {
	l.v = v
}

func (l *gaeLogger) Level() Lvl {
	return l.v
}

func (l *gaeLogger) Print(i ...interface{}) {
	l.Printf("%s", fmt.Sprintln(i...))
}

func (l *gaeLogger) Printf(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func (l *gaeLogger) Printj(j JSON) {
	l.Printf("%s", l.json(j))
}

func (l *gaeLogger) Debug(i ...interface{}) {
	l.Debugf("%s", fmt.Sprintln(i...))
}

func (l *gaeLogger) Debugf(format string, args ...interface{}) {
	log.Debugf(l.ctx, format, args...)
}

func (l *gaeLogger) Debugj(j JSON) {
	l.Debugf("%s", l.json(j))
}

func (l *gaeLogger) Info(i ...interface{}) {
	l.Infof("%s", fmt.Sprintln(i...))
}

func (l *gaeLogger) Infof(format string, args ...interface{}) {
	log.Infof(l.ctx, format, args...)
}

func (l *gaeLogger) Infoj(j JSON) {
	l.Infof("%s", l.json(j))
}

func (l *gaeLogger) Warn(i ...interface{}) {
	l.Warnf("%s", fmt.Sprintln(i...))
}

func (l *gaeLogger) Warnf(format string, args ...interface{}) {
	log.Warningf(l.ctx, format, args...)
}

func (l *gaeLogger) Warnj(j JSON) {
	l.Warnf("%s", l.json(j))
}

func (l *gaeLogger) Error(i ...interface{}) {
	l.Errorf("%s", fmt.Sprintln(i...))
}

func (l *gaeLogger) Errorf(format string, args ...interface{}) {
	log.Errorf(l.ctx, format, args...)
}

func (l *gaeLogger) Errorj(j JSON) {
	l.Errorf("%s", l.json(j))
}

func (l *gaeLogger) Fatal(i ...interface{}) {
	l.Fatalf("%s", fmt.Sprintln(i...))
}

func (l *gaeLogger) Fatalf(format string, args ...interface{}) {
	log.Criticalf(l.ctx, format, args...)
}

func (l *gaeLogger) Fatalj(j JSON) {
	l.Fatalf("%s", l.json(j))
}

func New(opts ...Option) Logger {
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}
	return &gaeLogger{
		ctx: opt.ctx,
		b:   new(bytes.Buffer),
	}
}
