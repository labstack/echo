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
		// TLS returns true if HTTP connection is TLS otherwise false.
		TLS() bool

		// Scheme returns the HTTP protocol scheme, `http` or `https`.
		Scheme() string

		// Host returns HTTP request host. // For server requests Host specifies the host on which the
		// URL is sought. Per RFC 2616, this is either the value of
		// the "Host" header or the host name given in the URL itself.
		Host() string

		// URI returns the unmodified `Request-URI` sent by the client.
		URI() string

		// URL returns `engine.URL`.
		URL() URL

		// Header returns `engine.Header`.
		Header() Header

		// Proto() string
		// ProtoMajor() int
		// ProtoMinor() int

		// UserAgent returns the client's `User-Agent`.
		UserAgent() string

		// RemoteAddress returns the client's network address.
		RemoteAddress() string

		// Method returns the HTTP method of the request.
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
		Keys() []string
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
