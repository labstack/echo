/*
Package echo implements a fast and unfancy micro web framework for Go.

Example:

    package main

    import (
        "net/http"

        "github.com/labstack/echo"
        mw "github.com/labstack/echo/middleware"
    )

    func hello(c *echo.Context) error {
        return c.String(http.StatusOK, "Hello, World!\n")
    }

    func main() {
        e := echo.New()

        e.Use(mw.Logger())
        e.Use(mw.Recover())

        e.Get("/", hello)

        e.Run(":1323")
    }

Learn more at https://labstack.com/echo
*/
package echo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"

	"github.com/labstack/gommon/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/websocket"
)

type (
	// Echo is the top-level framework instance.
	Echo struct {
		prefix           string
		middleware       []MiddlewareFunc
		http2            bool
		maxParam         *int
		httpErrorHandler HTTPErrorHandler
		binder           Binder
		renderer         Renderer
		pool             sync.Pool
		debug            bool
		hook             http.HandlerFunc
		autoIndex        bool
		logger           Logger
		router           *Router
	}

	// Logger is the interface that declares echo's logging system.
	Logger interface {
		Debug(...interface{})
		Debugf(string, ...interface{})

		Info(...interface{})
		Infof(string, ...interface{})

		Warn(...interface{})
		Warnf(string, ...interface{})

		Error(...interface{})
		Errorf(string, ...interface{})

		Fatal(...interface{})
		Fatalf(string, ...interface{})
	}

	// Route contains a handler and information for matching against requests.
	Route struct {
		Method  string
		Path    string
		Handler Handler
	}

	// HTTPError represents an error that occured while handling a request.
	HTTPError struct {
		code    int
		message string
	}

	// Middleware ...
	Middleware interface{}

	// MiddlewareFunc ...
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// Handler ...
	Handler interface{}

	// HandlerFunc ...
	HandlerFunc func(*Context) error

	// HTTPErrorHandler is a centralized HTTP error handler.
	HTTPErrorHandler func(error, *Context)

	// Validator is the interface that wraps the Validate method.
	Validator interface {
		Validate() error
	}

	// Renderer is the interface that wraps the Render method.
	Renderer interface {
		Render(w io.Writer, name string, data interface{}) error
	}
)

const (
	// CONNECT HTTP method
	CONNECT = "CONNECT"
	// DELETE HTTP method
	DELETE = "DELETE"
	// GET HTTP method
	GET = "GET"
	// HEAD HTTP method
	HEAD = "HEAD"
	// OPTIONS HTTP method
	OPTIONS = "OPTIONS"
	// PATCH HTTP method
	PATCH = "PATCH"
	// POST HTTP method
	POST = "POST"
	// PUT HTTP method
	PUT = "PUT"
	// TRACE HTTP method
	TRACE = "TRACE"

	//-------------
	// Media types
	//-------------

	ApplicationJSON                  = "application/json"
	ApplicationJSONCharsetUTF8       = ApplicationJSON + "; " + CharsetUTF8
	ApplicationJavaScript            = "application/javascript"
	ApplicationJavaScriptCharsetUTF8 = ApplicationJavaScript + "; " + CharsetUTF8
	ApplicationXML                   = "application/xml"
	ApplicationXMLCharsetUTF8        = ApplicationXML + "; " + CharsetUTF8
	ApplicationForm                  = "application/x-www-form-urlencoded"
	ApplicationProtobuf              = "application/protobuf"
	ApplicationMsgpack               = "application/msgpack"
	TextHTML                         = "text/html"
	TextHTMLCharsetUTF8              = TextHTML + "; " + CharsetUTF8
	TextPlain                        = "text/plain"
	TextPlainCharsetUTF8             = TextPlain + "; " + CharsetUTF8
	MultipartForm                    = "multipart/form-data"

	//---------
	// Charset
	//---------

	CharsetUTF8 = "charset=utf-8"

	//---------
	// Headers
	//---------

	AcceptEncoding     = "Accept-Encoding"
	Authorization      = "Authorization"
	ContentDisposition = "Content-Disposition"
	ContentEncoding    = "Content-Encoding"
	ContentLength      = "Content-Length"
	ContentType        = "Content-Type"
	Location           = "Location"
	Upgrade            = "Upgrade"
	Vary               = "Vary"
	WWWAuthenticate    = "WWW-Authenticate"
	XForwardedFor      = "X-Forwarded-For"
	XRealIP            = "X-Real-IP"
	//-----------
	// Protocols
	//-----------

	WebSocket = "websocket"

	indexPage = "index.html"
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

	//--------
	// Errors
	//--------

	ErrUnsupportedMediaType  = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrRendererNotRegistered = errors.New("renderer not registered")
	ErrInvalidRedirectCode   = errors.New("invalid redirect status code")

	//----------------
	// Error handlers
	//----------------

	notFoundHandler = func(c *Context) error {
		return NewHTTPError(http.StatusNotFound)
	}

	methodNotAllowedHandler = func(c *Context) error {
		return NewHTTPError(http.StatusMethodNotAllowed)
	}
)

// New creates an instance of Echo.
func New() (e *Echo) {
	e = &Echo{maxParam: new(int)}
	e.pool.New = func() interface{} {
		return NewContext(nil, new(Response), e)
	}
	e.router = NewRouter(e)

	//----------
	// Defaults
	//----------

	e.HTTP2(true)
	e.SetHTTPErrorHandler(e.DefaultHTTPErrorHandler)
	e.SetBinder(&binder{})

	// Logger
	e.logger = log.New("echo")

	return
}

// Router returns router.
func (e *Echo) Router() *Router {
	return e.router
}

// SetLogger sets the logger instance.
func (e *Echo) SetLogger(logger Logger) {
	e.logger = logger
}

// Logger returns the logger instance.
func (e *Echo) Logger() Logger {
	return e.logger
}

// HTTP2 enable/disable HTTP2 support.
func (e *Echo) HTTP2(on bool) {
	e.http2 = on
}

// DefaultHTTPErrorHandler invokes the default HTTP error handler.
func (e *Echo) DefaultHTTPErrorHandler(err error, c *Context) {
	code := http.StatusInternalServerError
	msg := http.StatusText(code)
	if he, ok := err.(*HTTPError); ok {
		code = he.code
		msg = he.message
	}
	if e.debug {
		msg = err.Error()
	}
	if !c.response.committed {
		http.Error(c.response, msg, code)
	}
	e.logger.Error(err)
}

// SetHTTPErrorHandler registers a custom Echo.HTTPErrorHandler.
func (e *Echo) SetHTTPErrorHandler(h HTTPErrorHandler) {
	e.httpErrorHandler = h
}

// SetBinder registers a custom binder. It's invoked by Context.Bind().
func (e *Echo) SetBinder(b Binder) {
	e.binder = b
}

// SetRenderer registers an HTML template renderer. It's invoked by Context.Render().
func (e *Echo) SetRenderer(r Renderer) {
	e.renderer = r
}

// SetDebug enable/disable debug mode.
func (e *Echo) SetDebug(on bool) {
	e.debug = on
}

// Debug returns debug mode (enabled or disabled).
func (e *Echo) Debug() bool {
	return e.debug
}

// AutoIndex enable/disable automatically creating an index page for the directory.
func (e *Echo) AutoIndex(on bool) {
	e.autoIndex = on
}

// Hook registers a callback which is invoked from `Echo#ServerHTTP` as the first
// statement. Hook is useful if you want to modify response/response objects even
// before it hits the router or any middleware.
func (e *Echo) Hook(h http.HandlerFunc) {
	e.hook = h
}

// Use adds handler to the middleware chain.
func (e *Echo) Use(m ...Middleware) {
	for _, h := range m {
		e.middleware = append(e.middleware, wrapMiddleware(h))
	}
}

// Connect adds a CONNECT route > handler to the router.
func (e *Echo) Connect(path string, h Handler) {
	e.add(CONNECT, path, h)
}

// Delete adds a DELETE route > handler to the router.
func (e *Echo) Delete(path string, h Handler) {
	e.add(DELETE, path, h)
}

// Get adds a GET route > handler to the router.
func (e *Echo) Get(path string, h Handler) {
	e.add(GET, path, h)
}

// Head adds a HEAD route > handler to the router.
func (e *Echo) Head(path string, h Handler) {
	e.add(HEAD, path, h)
}

// Options adds an OPTIONS route > handler to the router.
func (e *Echo) Options(path string, h Handler) {
	e.add(OPTIONS, path, h)
}

// Patch adds a PATCH route > handler to the router.
func (e *Echo) Patch(path string, h Handler) {
	e.add(PATCH, path, h)
}

// Post adds a POST route > handler to the router.
func (e *Echo) Post(path string, h Handler) {
	e.add(POST, path, h)
}

// Put adds a PUT route > handler to the router.
func (e *Echo) Put(path string, h Handler) {
	e.add(PUT, path, h)
}

// Trace adds a TRACE route > handler to the router.
func (e *Echo) Trace(path string, h Handler) {
	e.add(TRACE, path, h)
}

// Any adds a route > handler to the router for all HTTP methods.
func (e *Echo) Any(path string, h Handler) {
	for _, m := range methods {
		e.add(m, path, h)
	}
}

// Match adds a route > handler to the router for multiple HTTP methods provided.
func (e *Echo) Match(methods []string, path string, h Handler) {
	for _, m := range methods {
		e.add(m, path, h)
	}
}

// WebSocket adds a WebSocket route > handler to the router.
func (e *Echo) WebSocket(path string, h HandlerFunc) {
	e.Get(path, func(c *Context) (err error) {
		wss := websocket.Server{
			Handler: func(ws *websocket.Conn) {
				c.socket = ws
				c.response.status = http.StatusSwitchingProtocols
				err = h(c)
			},
		}
		wss.ServeHTTP(c.response, c.request)
		return err
	})
}

func (e *Echo) add(method, path string, h Handler) {
	path = e.prefix + path
	e.router.Add(method, path, wrapHandler(h), e)
	r := Route{
		Method:  method,
		Path:    path,
		Handler: runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name(),
	}
	e.router.routes = append(e.router.routes, r)
}

// Index serves index file.
func (e *Echo) Index(file string) {
	e.ServeFile("/", file)
}

// Favicon serves the default favicon - GET /favicon.ico.
func (e *Echo) Favicon(file string) {
	e.ServeFile("/favicon.ico", file)
}

// Static serves static files from a directory. It's an alias for `Echo.ServeDir`
func (e *Echo) Static(path, dir string) {
	e.ServeDir(path, dir)
}

// ServeDir serves files from a directory.
func (e *Echo) ServeDir(path, dir string) {
	e.Get(path+"*", func(c *Context) error {
		return e.serveFile(dir, c.P(0), c) // Param `_*`
	})
}

// ServeFile serves a file.
func (e *Echo) ServeFile(path, file string) {
	e.Get(path, func(c *Context) error {
		dir, file := filepath.Split(file)
		return e.serveFile(dir, file, c)
	})
}

func (e *Echo) serveFile(dir, file string, c *Context) (err error) {
	fs := http.Dir(dir)
	f, err := fs.Open(file)
	if err != nil {
		return NewHTTPError(http.StatusNotFound)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		/* NOTE:
		Not checking the Last-Modified header as it caches the response `304` when
		changing differnt directories for the same path.
		*/
		d := f

		// Index file
		file = path.Join(file, indexPage)
		f, err = fs.Open(file)
		if err != nil {
			if e.autoIndex {
				// Auto index
				return listDir(d, c)
			}
			return NewHTTPError(http.StatusForbidden)
		}
		fi, _ = f.Stat() // Index file stat
	}

	http.ServeContent(c.response, c.request, fi.Name(), fi.ModTime(), f)
	return
}

func listDir(d http.File, c *Context) (err error) {
	dirs, err := d.Readdir(-1)
	if err != nil {
		return err
	}

	// Create directory index
	w := c.Response()
	w.Header().Set(ContentType, TextHTMLCharsetUTF8)
	fmt.Fprintf(w, "<pre>\n")
	for _, d := range dirs {
		name := d.Name()
		color := "#212121"
		if d.IsDir() {
			color = "#e91e63"
			name += "/"
		}
		fmt.Fprintf(w, "<a href=\"%s\" style=\"color: %s;\">%s</a>\n", name, color, name)
	}
	fmt.Fprintf(w, "</pre>\n")
	return
}

// Group creates a new sub router with prefix. It inherits all properties from
// the parent. Passing middleware overrides parent middleware.
func (e *Echo) Group(prefix string, m ...Middleware) *Group {
	g := &Group{*e}
	g.echo.prefix += prefix
	if len(m) == 0 {
		mw := make([]MiddlewareFunc, len(g.echo.middleware))
		copy(mw, g.echo.middleware)
		g.echo.middleware = mw
	} else {
		g.echo.middleware = nil
		g.Use(m...)
	}
	return g
}

// URI generates a URI from handler.
func (e *Echo) URI(h Handler, params ...interface{}) string {
	uri := new(bytes.Buffer)
	pl := len(params)
	n := 0
	hn := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	for _, r := range e.router.routes {
		if r.Handler == hn {
			for i, l := 0, len(r.Path); i < l; i++ {
				if r.Path[i] == ':' && n < pl {
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
func (e *Echo) URL(h Handler, params ...interface{}) string {
	return e.URI(h, params...)
}

// Routes returns the registered routes.
func (e *Echo) Routes() []Route {
	return e.router.routes
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e.hook != nil {
		e.hook(w, r)
	}

	c := e.pool.Get().(*Context)
	h, e := e.router.Find(r.Method, r.URL.Path, c)
	c.reset(r, w, e)

	// Chain middleware with handler in the end
	for i := len(e.middleware) - 1; i >= 0; i-- {
		h = e.middleware[i](h)
	}

	// Execute chain
	if err := h(c); err != nil {
		e.httpErrorHandler(err, c)
	}

	e.pool.Put(c)
}

// Server returns the internal *http.Server.
func (e *Echo) Server(addr string) *http.Server {
	s := &http.Server{Addr: addr, Handler: e}
	// TODO: Remove in Go 1.6+
	if e.http2 {
		http2.ConfigureServer(s, nil)
	}
	return s
}

// Run runs a server.
func (e *Echo) Run(addr string) {
	e.run(e.Server(addr))
}

// RunTLS runs a server with TLS configuration.
func (e *Echo) RunTLS(addr, certfile, keyfile string) {
	e.run(e.Server(addr), certfile, keyfile)
}

// RunServer runs a custom server.
func (e *Echo) RunServer(s *http.Server) {
	e.run(s)
}

// RunTLSServer runs a custom server with TLS configuration.
func (e *Echo) RunTLSServer(s *http.Server, crtFile, keyFile string) {
	e.run(s, crtFile, keyFile)
}

func (e *Echo) run(s *http.Server, files ...string) {
	s.Handler = e
	// TODO: Remove in Go 1.6+
	if e.http2 {
		http2.ConfigureServer(s, nil)
	}
	if len(files) == 0 {
		e.logger.Fatal(s.ListenAndServe())
	} else if len(files) == 2 {
		e.logger.Fatal(s.ListenAndServeTLS(files[0], files[1]))
	} else {
		e.logger.Fatal("invalid TLS configuration")
	}
}

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int, msg ...string) *HTTPError {
	he := &HTTPError{code: code, message: http.StatusText(code)}
	if len(msg) > 0 {
		m := msg[0]
		he.message = m
	}
	return he
}

// SetCode sets code.
func (e *HTTPError) SetCode(code int) {
	e.code = code
}

// Code returns code.
func (e *HTTPError) Code() int {
	return e.code
}

// Error returns message.
func (e *HTTPError) Error() string {
	return e.message
}

// wrapMiddleware wraps middleware.
func wrapMiddleware(m Middleware) MiddlewareFunc {
	switch m := m.(type) {
	case MiddlewareFunc:
		return m
	case func(HandlerFunc) HandlerFunc:
		return m
	case HandlerFunc:
		return wrapHandlerFuncMW(m)
	case func(*Context) error:
		return wrapHandlerFuncMW(m)
	case func(http.Handler) http.Handler:
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) (err error) {
				m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					c.response.writer = w
					c.request = r
					err = h(c)
				})).ServeHTTP(c.response.writer, c.request)
				return
			}
		}
	case http.Handler:
		return wrapHTTPHandlerFuncMW(m.ServeHTTP)
	case func(http.ResponseWriter, *http.Request):
		return wrapHTTPHandlerFuncMW(m)
	default:
		panic("unknown middleware")
	}
}

// wrapHandlerFuncMW wraps HandlerFunc middleware.
func wrapHandlerFuncMW(m HandlerFunc) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			if err := m(c); err != nil {
				return err
			}
			return h(c)
		}
	}
}

// wrapHTTPHandlerFuncMW wraps http.HandlerFunc middleware.
func wrapHTTPHandlerFuncMW(m http.HandlerFunc) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			if !c.response.committed {
				m.ServeHTTP(c.response.writer, c.request)
			}
			return h(c)
		}
	}
}

// wrapHandler wraps handler.
func wrapHandler(h Handler) HandlerFunc {
	switch h := h.(type) {
	case HandlerFunc:
		return h
	case func(*Context) error:
		return h
	case http.Handler, http.HandlerFunc:
		return func(c *Context) error {
			h.(http.Handler).ServeHTTP(c.response, c.request)
			return nil
		}
	case func(http.ResponseWriter, *http.Request):
		return func(c *Context) error {
			h(c.response, c.request)
			return nil
		}
	default:
		panic("unknown handler")
	}
}
