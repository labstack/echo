package engine

import (
	"io"
	"time"

	"github.com/labstack/gommon/log"
)

type (
	// Engine defines an interface for HTTP server.
	Engine interface {
		SetHandler(Handler)
		SetLogger(*log.Logger)
		Start()
	}

	// Request defines an interface for HTTP request.
	Request interface {
		TLS() bool
		Scheme() string
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
	}

	// Response defines an interface for HTTP response.
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

	// Header defines an interface for HTTP header.
	Header interface {
		Add(string, string)
		Del(string)
		Get(string) string
		Set(string, string)
	}

	// URL defines an interface for HTTP request url.
	URL interface {
		SetPath(string)
		Path() string
		QueryValue(string) string
	}

	// Config defines engine configuration.
	Config struct {
		Address      string
		TLSCertfile  string
		TLSKeyfile   string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}

	// Handler defines an interface to server HTTP requests via `ServeHTTP(Request, Response)`
	// function.
	Handler interface {
		ServeHTTP(Request, Response)
	}

	// HandlerFunc is an adapter to allow the use of `func(Request, Response)` as HTTP handlers.
	HandlerFunc func(Request, Response)
)

// ServeHTTP serves HTTP request.
func (h HandlerFunc) ServeHTTP(req Request, res Response) {
	h(req, res)
}
