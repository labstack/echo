package engine

import (
	"io"
	"time"
)

type (
	Type uint8

	HandlerFunc func(Request, Response)

	Engine interface {
		Start()
	}

	Request interface {
		Header() Header
		// Proto() string
		// ProtoMajor() int
		// ProtoMinor() int
		RemoteAddress() string
		Method() string
		URI() string
		URL() URL
		Body() io.ReadCloser
		FormValue(string) string
	}

	Response interface {
		Header() Header
		WriteHeader(int)
		Write(b []byte) (int, error)
		Status() int
		Size() int64
		Committed() bool
		SetWriter(io.Writer)
		Writer() io.Writer
	}

	Header interface {
		Add(string, string)
		Del(string)
		Get(string) string
		Set(string, string)
	}

	URL interface {
		Scheme() string
		SetPath(string)
		Path() string
		Host() string
		QueryValue(string) string
	}

	Config struct {
		Address      string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		TLSCertfile  string
		TLSKeyfile   string
	}
)

const (
	Standard Type = iota
	FastHTTP
)
