package echo

import (
	"bytes"
	"io"
	"strconv"
	"sync"
	"time"
)

//-----------------------------------------------------------------------------
// Example for Zap (https://github.com/uber-go/zap)
//func main() {
//	e := echo.New()
//	logger, _ := zap.NewProduction()
//	e.Logger = &ZapLogger{logger: logger}
//}
//type ZapLogger struct {
//	logger *zap.Logger
//}
//
//func (l *ZapLogger) Write(p []byte) (n int, err error) {
//	// Note: if `logger` middleware is used it will send json bytes here, and it will not look beautiful at all.
//	l.logger.Info(string(p), zap.String("subsystem", "echo")) // naively log everything as string message.
//	return len(p), nil
//}
//
//func (l *ZapLogger) Error(err error) {
//	l.logger.Error(err.Error(), zap.Error(err), zap.String("subsystem", "echo"))
//}

//-----------------------------------------------------------------------------
// Example for Zerolog (https://github.com/rs/zerolog)
//func main() {
//	e := echo.New()
//	logger := zerolog.New(os.Stdout)
//	e.Logger = &ZeroLogger{logger: &logger}
//}
//
//type ZeroLogger struct {
//	logger *zerolog.Logger
//}
//
//func (l *ZeroLogger) Write(p []byte) (n int, err error) {
//	// Note: if `logger` middleware is used it will send json bytes here, and it will not look beautiful at all.
//	l.logger.Info().Str("subsystem", "echo").Msg(string(p)) // naively log everything as string message.
//	return len(p), nil
//}
//
//func (l *ZeroLogger) Error(err error) {
//	l.logger.Error().Str("subsystem", "echo").Err(err).Msg(err.Error())
//}

//-----------------------------------------------------------------------------
// Example for Logrus (https://github.com/sirupsen/logrus)
//func main() {
//	e := echo.New()
//	e.Logger = &LogrusLogger{logger: logrus.New()}
//}
//
//type LogrusLogger struct {
//	logger *logrus.Logger
//}
//
//func (l *LogrusLogger) Write(p []byte) (n int, err error) {
//	// Note: if `logger` middleware is used it will send json bytes here, and it will not look beautiful at all.
//	l.logger.WithFields(logrus.Fields{"subsystem": "echo"}).Info(string(p)) // naively log everything as string message.
//	return len(p), nil
//}
//
//func (l *LogrusLogger) Error(err error) {
//	l.logger.WithFields(logrus.Fields{"subsystem": "echo"}).Error(err)
//}

// Logger defines the logging interface that Echo uses internally in few places.
// For logging in handlers use your own logger instance (dependency injected or package/public variable) from logging framework of your choice.
type Logger interface {
	// Write provides writer interface for http.Server `ErrorLog` and for logging startup messages.
	// `http.Server.ErrorLog` logs errors from accepting connections, unexpected behavior from handlers,
	// and underlying FileSystem errors.
	// `logger` middleware will use this method to write its JSON payload.
	Write(p []byte) (n int, err error)
	// Error logs the error
	Error(err error)
}

// jsonLogger is similar logger formatting implementation as `v4` had. It is not particularly fast or efficient. Only
// goal it to exist is to have somewhat backwards compatibility with `v4` for Echo internals logging formatting.
// It is not meant for logging in handlers/middlewares. Use some real logging library for those cases.
type jsonLogger struct {
	writer     io.Writer
	bufferPool sync.Pool
	lock       sync.Mutex

	timeNow func() time.Time
}

func newJSONLogger(writer io.Writer) *jsonLogger {
	return &jsonLogger{
		writer: writer,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 256))
			},
		},
		timeNow: time.Now,
	}
}

func (l *jsonLogger) Write(p []byte) (n int, err error) {
	pLen := len(p)
	if pLen >= 2 && // naively try to avoid JSON values to be wrapped into message
		(p[0] == '{' && p[pLen-2] == '}' && p[pLen-1] == '\n') ||
		(p[0] == '{' && p[pLen-1] == '}') {
		return l.write(p)
	}
	// we log with WARN level as we have no idea what that message level should be. From Echo perspective this method is
	// called when we pass Echo logger to http.Server.ErrorLog and there are problems inside http.Server - which probably
	// deserves at least WARN level.
	return l.printf("INFO", string(p))
}

func (l *jsonLogger) Error(err error) {
	_, _ = l.printf("ERROR", err.Error())
}

func (l *jsonLogger) printf(level string, message string) (n int, err error) {
	buf := l.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer l.bufferPool.Put(buf)

	buf.WriteString(`{"time":"`)
	buf.WriteString(l.timeNow().Format(time.RFC3339Nano))
	buf.WriteString(`","level":"`)
	buf.WriteString(level)
	buf.WriteString(`","prefix":"echo","message":`)

	buf.WriteString(strconv.Quote(message))
	buf.WriteString("}\n")

	return l.write(buf.Bytes())
}

func (l *jsonLogger) write(p []byte) (int, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.writer.Write(p)
}
