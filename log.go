package echo

import (
	"io"
	"time"

	"github.com/labstack/gommon/log"
)

type (
	// Logger defines the logging interface.
	Logger interface {
		Output() io.Writer
		SetOutput(w io.Writer)
		Prefix() string
		SetPrefix(p string)
		Level() log.Lvl
		SetLevel(v log.Lvl)
		Print(i ...interface{})
		Printf(format string, args ...interface{})
		Printj(j log.JSON)
		Debug(i ...interface{})
		Debugf(format string, args ...interface{})
		Debugj(j log.JSON)
		Info(i ...interface{})
		Infof(format string, args ...interface{})
		Infoj(j log.JSON)
		Warn(i ...interface{})
		Warnf(format string, args ...interface{})
		Warnj(j log.JSON)
		Error(i ...interface{})
		Errorf(format string, args ...interface{})
		Errorj(j log.JSON)
		Fatal(i ...interface{})
		Fatalj(j log.JSON)
		Fatalf(format string, args ...interface{})
		Panic(i ...interface{})
		Panicj(j log.JSON)
		Panicf(format string, args ...interface{})
	}

	// RequestLogger defines the logging interface for logging middleware.
	RequestLogger interface {
		LogRequest(*Request) error
	}

	// Request defines the data to be logged by logger middleware.
	Request struct {
		// ID string `json:"id,omitempty"` (Request ID - Not implemented)
		Time         int64             `json:"time,omitempty"`
		RemoteIP     string            `json:"remote_ip,omitempty"`
		URI          string            `json:"uri,omitempty"`
		Host         string            `json:"host,omitempty"`
		Method       string            `json:"method,omitempty"`
		Path         string            `json:"path,omitempty"`
		Referer      string            `json:"referer,omitempty"`
		UserAgent    string            `json:"user_agent,omitempty"`
		Status       int               `json:"status,omitempty"`
		Latency      time.Duration     `json:"latency,omitempty"`
		LatencyHuman string            `json:"latency_human,omitempty"`
		BytesIn      int64             `json:"bytes_in"`
		BytesOut     int64             `json:"bytes_out"`
		Header       map[string]string `json:"header,omitempty"`
		Form         map[string]string `json:"form,omitempty"`
		Query        map[string]string `json:"query,omitempty"`
	}
)
