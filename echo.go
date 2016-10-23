/*
Package echo implements a fast and unfancy HTTP server framework for Go (Golang).

Example:

	package main

	import (
	    "net/http"

	    "github.com/labstack/echo"
	    "github.com/labstack/echo/engine/standard"
	    "github.com/labstack/echo/middleware"
	)

	// Handler
	func hello(c echo.Context) error {
	    return c.String(http.StatusOK, "Hello, World!")
	}

	func main() {
	    // Echo instance
	    e := echo.New()

	    // Middleware
	    e.Use(middleware.Logger())
	    e.Use(middleware.Recover())

	    // Routes
	    e.GET("/", hello)

	    // Start server
	    if err := e.Start(":1323"); err != nil {
			e.Logger.Fatal(err)
		}
	}

Learn more at https://echo.labstack.com
*/
package echo

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	slog "log"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/rsc/letsencrypt"

	"github.com/labstack/gommon/color"
	glog "github.com/labstack/gommon/log"
	"github.com/tylerb/graceful"
)

type (
	// Echo is the top-level framework instance.
	Echo struct {
		Server          *http.Server
		TLSServer       *http.Server
		TLSConfig       *tls.Config
		TLSCacheFile    string
		TLSHosts        []string
		ShutdownTimeout time.Duration
		DisableHTTP2    bool
		Debug           bool
		HTTPErrorHandler
		Binder          Binder
		Renderer        Renderer
		Logger          Logger
		tlsManager      letsencrypt.Manager
		graceful        *graceful.Server
		gracefulTLS     *graceful.Server
		premiddleware   []MiddlewareFunc
		middleware      []MiddlewareFunc
		maxParam        *int
		router          *Router
		notFoundHandler HandlerFunc
		pool            sync.Pool
		color           *color.Color
	}

	// Route contains a handler and information for matching against requests.
	Route struct {
		Method  string
		Path    string
		Handler string
	}

	// HTTPError represents an error that occurred while handling a request.
	HTTPError struct {
		Code    int
		Message string
	}

	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// HandlerFunc defines a function to server HTTP requests.
	HandlerFunc func(Context) error

	// HTTPErrorHandler is a centralized HTTP error handler.
	HTTPErrorHandler func(error, Context)

	// Validator is the interface that wraps the Validate function.
	Validator interface {
		Validate() error
	}

	// Renderer is the interface that wraps the Render function.
	Renderer interface {
		Render(io.Writer, string, interface{}, Context) error
	}

	// Map defines a generic map of type `map[string]interface{}`.
	Map map[string]interface{}
)

// HTTP methods
const (
	CONNECT = "CONNECT"
	DELETE  = "DELETE"
	GET     = "GET"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	POST    = "POST"
	PUT     = "PUT"
	TRACE   = "TRACE"
)

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

const (
	charsetUTF8 = "charset=utf-8"
)

// Headers
const (
	HeaderAcceptEncoding                = "Accept-Encoding"
	HeaderAllow                         = "Allow"
	HeaderAuthorization                 = "Authorization"
	HeaderContentDisposition            = "Content-Disposition"
	HeaderContentEncoding               = "Content-Encoding"
	HeaderContentLength                 = "Content-Length"
	HeaderContentType                   = "Content-Type"
	HeaderCookie                        = "Cookie"
	HeaderSetCookie                     = "Set-Cookie"
	HeaderIfModifiedSince               = "If-Modified-Since"
	HeaderLastModified                  = "Last-Modified"
	HeaderLocation                      = "Location"
	HeaderUpgrade                       = "Upgrade"
	HeaderVary                          = "Vary"
	HeaderWWWAuthenticate               = "WWW-Authenticate"
	HeaderXForwardedProto               = "X-Forwarded-Proto"
	HeaderXHTTPMethodOverride           = "X-HTTP-Method-Override"
	HeaderXForwardedFor                 = "X-Forwarded-For"
	HeaderXRealIP                       = "X-Real-IP"
	HeaderServer                        = "Server"
	HeaderOrigin                        = "Origin"
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)

var (
	methods = [...]string{
		CONNECT,
		DELETE,
		GET,
		HEAD,
		OPTIONS,
		PATCH,
		POST,
		PUT,
		TRACE,
	}
)

// Errors
var (
	ErrUnsupportedMediaType        = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrNotFound                    = NewHTTPError(http.StatusNotFound)
	ErrUnauthorized                = NewHTTPError(http.StatusUnauthorized)
	ErrMethodNotAllowed            = NewHTTPError(http.StatusMethodNotAllowed)
	ErrStatusRequestEntityTooLarge = NewHTTPError(http.StatusRequestEntityTooLarge)
	ErrRendererNotRegistered       = errors.New("renderer not registered")
	ErrInvalidRedirectCode         = errors.New("invalid redirect status code")
	ErrCookieNotFound              = errors.New("cookie not found")
)

// Error handlers
var (
	NotFoundHandler = func(c Context) error {
		return ErrNotFound
	}

	MethodNotAllowedHandler = func(c Context) error {
		return ErrMethodNotAllowed
	}
)

// New creates an instance of Echo.

func New() (e *Echo) {
	e = &Echo{
		Server:    new(http.Server),
		TLSServer: new(http.Server),
		// TODO: https://github.com/golang/go/commit/d24f446a90ea94b87591bf16228d7d871fec3d92
		TLSConfig:       new(tls.Config),
		TLSCacheFile:    "letsencrypt.cache",
		ShutdownTimeout: 15 * time.Second,
		Logger:          glog.New("echo"),
		graceful:        new(graceful.Server),
		gracefulTLS:     new(graceful.Server),
		maxParam:        new(int),
		color:           color.New(),
	}
	e.HTTPErrorHandler = e.DefaultHTTPErrorHandler
	e.Binder = &binder{}
	e.Logger.SetLevel(glog.OFF)
	e.pool.New = func() interface{} {
		return e.NewContext(nil, nil)
	}
	e.router = NewRouter(e)
	return
}

// NewContext returns a Context instance.
func (e *Echo) NewContext(r *http.Request, w http.ResponseWriter) Context {
	return &context{
		request:  r,
		response: NewResponse(w, e),
		store:    make(Map),
		echo:     e,
		pvalues:  make([]string, *e.maxParam),
		handler:  NotFoundHandler,
	}
}

// Router returns router.
func (e *Echo) Router() *Router {
	return e.router
}

// DefaultHTTPErrorHandler invokes the default HTTP error handler.
func (e *Echo) DefaultHTTPErrorHandler(err error, c Context) {
	code := http.StatusInternalServerError
	msg := http.StatusText(code)
	if he, ok := err.(*HTTPError); ok {
		code = he.Code
		msg = he.Message
	}
	if e.Debug {
		msg = err.Error()
	}
	if !c.Response().Committed {
		if c.Request().Method == HEAD { // Issue #608
			c.NoContent(code)
		} else {
			c.String(code, msg)
		}
	}
	e.Logger.Error(err)
}

// Pre adds middleware to the chain which is run before router.
func (e *Echo) Pre(middleware ...MiddlewareFunc) {
	e.premiddleware = append(e.premiddleware, middleware...)
}

// Use adds middleware to the chain which is run after router.
func (e *Echo) Use(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Echo) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.add(CONNECT, path, h, m...)
}

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
func (e *Echo) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.add(DELETE, path, h, m...)
}

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (e *Echo) GET(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.add(GET, path, h, m...)
}

// HEAD registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Echo) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.add(HEAD, path, h, m...)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Echo) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.add(OPTIONS, path, h, m...)
}

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Echo) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.add(PATCH, path, h, m...)
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Echo) POST(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.add(POST, path, h, m...)
}

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Echo) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.add(PUT, path, h, m...)
}

// TRACE registers a new TRACE route for a path with matching handler in the
// router with optional route-level middleware.
func (e *Echo) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) {
	e.add(TRACE, path, h, m...)
}

// Any registers a new route for all HTTP methods and path with matching handler
// in the router with optional route-level middleware.
func (e *Echo) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	for _, m := range methods {
		e.add(m, path, handler, middleware...)
	}
}

// Match registers a new route for multiple HTTP methods and path with matching
// handler in the router with optional route-level middleware.
func (e *Echo) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	for _, m := range methods {
		e.add(m, path, handler, middleware...)
	}
}

// Static registers a new route with path prefix to serve static files from the
// provided root directory.
func (e *Echo) Static(prefix, root string) {
	e.GET(prefix+"*", func(c Context) error {
		return c.File(path.Join(root, c.Param("*")))
	})
}

// File registers a new route with path to serve a static file.
func (e *Echo) File(path, file string) {
	e.GET(path, func(c Context) error {
		return c.File(file)
	})
}

func (e *Echo) add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	name := handlerName(handler)
	e.router.Add(method, path, func(c Context) error {
		h := handler
		// Chain middleware
		for i := len(middleware) - 1; i >= 0; i-- {
			h = middleware[i](h)
		}
		return h(c)
	})
	r := Route{
		Method:  method,
		Path:    path,
		Handler: name,
	}
	e.router.routes[method+path] = r
}

// Group creates a new router group with prefix and optional group-level middleware.
func (e *Echo) Group(prefix string, m ...MiddlewareFunc) (g *Group) {
	g = &Group{prefix: prefix, echo: e}
	g.Use(m...)
	return
}

// URI generates a URI from handler.
func (e *Echo) URI(handler HandlerFunc, params ...interface{}) string {
	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0
	name := handlerName(handler)
	for _, r := range e.router.routes {
		if r.Handler == name {
			for i, l := 0, len(r.Path); i < l; i++ {
				if r.Path[i] == ':' && n < ln {
					for ; i < l && r.Path[i] != '/'; i++ {
					}
					uri.WriteString(fmt.Sprintf("%v", params[n]))
					n++
				}
				if i < l {
					uri.WriteByte(r.Path[i])
				}
			}
			break
		}
	}
	return uri.String()
}

// URL is an alias for `URI` function.
func (e *Echo) URL(h HandlerFunc, params ...interface{}) string {
	return e.URI(h, params...)
}

// Routes returns the registered routes.
func (e *Echo) Routes() []Route {
	routes := []Route{}
	for _, v := range e.router.routes {
		routes = append(routes, v)
	}
	return routes
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must be return the context by calling `ReleaseContext()`.
func (e *Echo) AcquireContext() Context {
	return e.pool.Get().(Context)
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (e *Echo) ReleaseContext(c Context) {
	e.pool.Put(c)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := e.pool.Get().(*context)
	c.Reset(r, w)

	// Middleware
	h := func(c Context) error {
		method := r.Method
		path := r.URL.Path
		e.router.Find(method, path, c)
		h := c.Handler()
		for i := len(e.middleware) - 1; i >= 0; i-- {
			h = e.middleware[i](h)
		}
		return h(c)
	}

	// Premiddleware
	for i := len(e.premiddleware) - 1; i >= 0; i-- {
		h = e.premiddleware[i](h)
	}

	// Execute chain
	if err := h(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}

	e.pool.Put(c)
}

func (e *Echo) config(gs *graceful.Server, s *http.Server, address string) {
	e.color.SetOutput(e.Logger.Output())
	gs.Server = s
	gs.Addr = address
	gs.Handler = e
	gs.Timeout = e.ShutdownTimeout
	gs.Logger = slog.New(e.Logger.Output(), e.Logger.Prefix()+": ", 0)
	if gs == e.gracefulTLS && !e.DisableHTTP2 {
		e.TLSConfig.NextProtos = append(e.TLSConfig.NextProtos, "h2")
	}
}

// Start starts the HTTP server.
// Note: If custom `Echo#Server` is used, it's Addr and Handler properties are
// ignored.
func (e *Echo) Start(address string) (err error) {
	e.config(e.graceful, e.Server, address)
	color.Printf(" ⇛ http server started on %s\n", color.Green(address))
	return e.graceful.ListenAndServe()
}

// StartTLS starts the TLS server.
// Note: If custom `Echo#TLSServer` is used, it's Addr and Handler properties are
// ignored.
func (e *Echo) StartTLS(address string, certFile, keyFile string) (err error) {
	e.config(e.gracefulTLS, e.TLSServer, address)
	if certFile == "" || keyFile == "" {
		return errors.New("invalid tls configuration")
	}
	e.TLSConfig.Certificates = make([]tls.Certificate, 1)
	e.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return
	}
	color.Printf(" ⇛ https server started on %s\n", color.Green(address))
	return e.gracefulTLS.ListenAndServeTLSConfig(e.TLSConfig)
}

// StartAutoTLS starts the TLS server using certificates automatically from https://letsencrypt.org.
// Note: If custom `Echo#TLSServer` is used, it's Addr and Handler properties are
// ignored.
func (e *Echo) StartAutoTLS() (err error) {
	address := ":443"
	e.config(e.gracefulTLS, e.TLSServer, address)
	e.TLSConfig.GetCertificate = e.tlsManager.GetCertificate
	if err = e.tlsManager.CacheFile(e.TLSCacheFile); err != nil {
		return
	}
	// NOTE: Added security
	e.tlsManager.SetHosts(e.TLSHosts)
	color.Printf(" ⇛ https auto tls server started on %s\n", color.Green(address))
	return e.gracefulTLS.ListenAndServeTLSConfig(e.TLSConfig)
}

// Shutdown gracefully shutdown the HTTP server with timeout.
func (e *Echo) Shutdown(timeout time.Duration) {
	e.graceful.Stop(timeout)
}

// ShutdownTLS gracefully shutdown the TLS server with timeout.
func (e *Echo) ShutdownTLS(timeout time.Duration) {
	e.gracefulTLS.Stop(timeout)
}

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int, msg ...string) *HTTPError {
	he := &HTTPError{Code: code, Message: http.StatusText(code)}
	if len(msg) > 0 {
		m := msg[0]
		he.Message = m
	}
	return he
}

// Error makes it compatible with `error` interface.
func (e *HTTPError) Error() string {
	return e.Message
}

// WrapHandler wraps `http.Handler` into `echo.HandlerFunc`.
func WrapHandler(h http.Handler) HandlerFunc {
	return func(c Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

// WrapMiddleware wraps `func(http.Handler) http.Handler` into `echo.MiddlewareFunc`
func WrapMiddleware(m func(http.Handler) http.Handler) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) (err error) {
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err = next(c)
			})).ServeHTTP(c.Response(), c.Request())
			return
		}
	}
}

func handlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}
