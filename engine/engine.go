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
		Host() string
		URI() string
		URL() URL
		Header() Header
		// Proto() string
		// ProtoMajor() int
		// ProtoMinor() int
		RemoteAddress() string
		Method() string
		Body() io.ReadCloser
		FormValue(string) string
		Object() interface{}
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
		Object() interface{}
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
