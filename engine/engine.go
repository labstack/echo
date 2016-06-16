package engine

import (
	"io"
	"mime/multipart"
	"time"

	"net"

	"github.com/labstack/echo/log"
)

type (
	// Server defines the interface for HTTP server.
	Server interface {
		// SetHandler sets the handler for the HTTP server.
		SetHandler(Handler)

		// SetLogger sets the logger for the HTTP server.
		SetLogger(log.Logger)

		// Start starts the HTTP server.
		Start() error
	}

	// Request defines the interface for HTTP request.
	Request interface {
		// IsTLS returns true if HTTP connection is TLS otherwise false.
		IsTLS() bool

		// Scheme returns the HTTP protocol scheme, `http` or `https`.
		Scheme() string

		// Host returns HTTP request host. Per RFC 2616, this is either the value of
		// the `Host` header or the host name given in the URL itself.
		Host() string

		// URI returns the unmodified `Request-URI` sent by the client.
		URI() string

		// SetURI sets the URI of the request.
		SetURI(string)

		// URL returns `engine.URL`.
		URL() URL

		// Header returns `engine.Header`.
		Header() Header

		// Referer returns the referring URL, if sent in the request.
		Referer() string

		// Protocol returns the protocol version string of the HTTP request.
		// Protocol() string

		// ProtocolMajor returns the major protocol version of the HTTP request.
		// ProtocolMajor() int

		// ProtocolMinor returns the minor protocol version of the HTTP request.
		// ProtocolMinor() int

		// ContentLength returns the size of request's body.
		ContentLength() int64

		// UserAgent returns the client's `User-Agent`.
		UserAgent() string

		// RemoteAddress returns the client's network address.
		RemoteAddress() string

		// Method returns the request's HTTP function.
		Method() string

		// SetMethod sets the HTTP method of the request.
		SetMethod(string)

		// Body returns request's body.
		Body() io.Reader

		// Body sets request's body.
		SetBody(io.Reader)

		// FormValue returns the form field value for the provided name.
		FormValue(string) string

		// FormParams returns the form parameters.
		FormParams() map[string][]string

		// FormFile returns the multipart form file for the provided name.
		FormFile(string) (*multipart.FileHeader, error)

		// MultipartForm returns the multipart form.
		MultipartForm() (*multipart.Form, error)

		// Cookie returns the named cookie provided in the request.
		Cookie(string) (Cookie, error)

		// Cookies returns the HTTP cookies sent with the request.
		Cookies() []Cookie
	}

	// Response defines the interface for HTTP response.
	Response interface {
		// Header returns `engine.Header`
		Header() Header

		// WriteHeader sends an HTTP response header with status code.
		WriteHeader(int)

		// Write writes the data to the connection as part of an HTTP reply.
		Write(b []byte) (int, error)

		// SetCookie adds a `Set-Cookie` header in HTTP response.
		SetCookie(Cookie)

		// Status returns the HTTP response status.
		Status() int

		// Size returns the number of bytes written to HTTP response.
		Size() int64

		// Committed returns true if HTTP response header is written, otherwise false.
		Committed() bool

		// Write returns the HTTP response writer.
		Writer() io.Writer

		// SetWriter sets the HTTP response writer.
		SetWriter(io.Writer)
	}

	// Header defines the interface for HTTP header.
	Header interface {
		// Add adds the key, value pair to the header. It appends to any existing values
		// associated with key.
		Add(string, string)

		// Del deletes the values associated with key.
		Del(string)

		// Set sets the header entries associated with key to the single element value.
		// It replaces any existing values associated with key.
		Set(string, string)

		// Get gets the first value associated with the given key. If there are
		// no values associated with the key, Get returns "".
		Get(string) string

		// Keys returns the header keys.
		Keys() []string

		// Contains checks if the header is set.
		Contains(string) bool
	}

	// URL defines the interface for HTTP request url.
	URL interface {
		// Path returns the request URL path.
		Path() string

		// SetPath sets the request URL path.
		SetPath(string)

		// QueryParam returns the query param for the provided name.
		QueryParam(string) string

		// QueryParam returns the query parameters as map.
		QueryParams() map[string][]string

		// QueryString returns the URL query string.
		QueryString() string
	}

	// Cookie defines the interface for HTTP cookie.
	Cookie interface {
		// Name returns the name of the cookie.
		Name() string

		// Value returns the value of the cookie.
		Value() string

		// Path returns the path of the cookie.
		Path() string

		// Domain returns the domain of the cookie.
		Domain() string

		// Expires returns the expiry time of the cookie.
		Expires() time.Time

		// Secure indicates if cookie is secured.
		Secure() bool

		// HTTPOnly indicate if cookies is HTTP only.
		HTTPOnly() bool
	}

	// Config defines engine config.
	Config struct {
		Address      string        // TCP address to listen on.
		Listener     net.Listener  // Custom `net.Listener`. If set, server accepts connections on it.
		TLSCertFile  string        // TLS certificate file path.
		TLSKeyFile   string        // TLS key file path.
		ReadTimeout  time.Duration // Maximum duration before timing out read of the request.
		WriteTimeout time.Duration // Maximum duration before timing out write of the response.
	}

	// Handler defines an interface to server HTTP requests via `ServeHTTP(Request, Response)`
	// function.
	Handler interface {
		ServeHTTP(Request, Response)
	}

	// HandlerFunc is an adapter to allow the use of `func(Request, Response)` as
	// an HTTP handler.
	HandlerFunc func(Request, Response)
)

// ServeHTTP serves HTTP request.
func (h HandlerFunc) ServeHTTP(req Request, res Response) {
	h(req, res)
}
