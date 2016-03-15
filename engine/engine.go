package engine

import (
	"io"
	"mime/multipart"
	"time"

	"github.com/labstack/gommon/log"
)

type (
	// Engine defines the interface for HTTP server.
	Engine interface {
		SetHandler(Handler)
		SetLogger(*log.Logger)
		Start()
	}

	// Request defines the interface for HTTP request.
	Request interface {
		// TLS returns true if connection is TLS otherwise false.
		TLS() bool

		Scheme() string
		Host() string
		URI() string
		URL() URL
		Header() Header
		// Proto() string
		// ProtoMajor() int
		// ProtoMinor() int
		UserAgent() string
		RemoteAddress() string
		Method() string
		SetMethod(string)
		Body() io.ReadCloser
		FormValue(string) string
		FormFile(string) (*multipart.FileHeader, error)
		MultipartForm() (*multipart.Form, error)
	}

	// Response defines the interface for HTTP response.
	Response interface {
		Header() Header
		WriteHeader(int)
		Write(b []byte) (int, error)
		Status() int
		Size() int64
		Committed() bool
		Writer() io.Writer
		SetWriter(io.Writer)
	}

	// Header defines the interface for HTTP header.
	Header interface {
		Add(string, string)
		Del(string)
		Set(string, string)
		Get(string) string
	}

	// URL defines the interface for HTTP request url.
	URL interface {
		Path() string
		SetPath(string)
		QueryValue(string) string
		QueryString() string
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
