package echo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/bradfitz/http2"
	"github.com/mattn/go-colorable"
	"golang.org/x/net/websocket"
)

type (
	Echo struct {
		router           *Router
		prefix           string
		middleware       []MiddlewareFunc
		http2            bool
		maxParam         byte
		notFoundHandler  HandlerFunc
		httpErrorHandler HTTPErrorHandler
		binder           BindFunc
		renderer         Renderer
		uris             map[Handler]string
		pool             sync.Pool
		debug            bool
	}

	HTTPError struct {
		code    int
		message string
	}

	Middleware     interface{}
	MiddlewareFunc func(HandlerFunc) HandlerFunc
	Handler        interface{}
	HandlerFunc    func(*Context) error

	// HTTPErrorHandler is a centralized HTTP error handler.
	HTTPErrorHandler func(error, *Context)

	BindFunc func(*http.Request, interface{}) error

	// Renderer is the interface that wraps the Render method.
	//
	// Render renders the HTML template with given name and specified data.
	// It writes the output to w.
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

	ApplicationJSON     = "application/json"
	ApplicationProtobuf = "application/protobuf"
	ApplicationMsgpack  = "application/msgpack"
	TextPlain           = "text/plain"
	TextHTML            = "text/html"
	ApplicationForm     = "application/x-www-form-urlencoded"
	MultipartForm       = "multipart/form-data"

	//---------
	// Headers
	//---------

	Accept             = "Accept"
	AcceptEncoding     = "Accept-Encoding"
	ContentDisposition = "Content-Disposition"
	ContentEncoding    = "Content-Encoding"
	ContentLength      = "Content-Length"
	ContentType        = "Content-Type"
	Authorization      = "Authorization"
	Upgrade            = "Upgrade"

	//-----------
	// Protocols
	//-----------

	WebSocket = "websocket"
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

	UnsupportedMediaType  = errors.New("echo ⇒ unsupported media type")
	RendererNotRegistered = errors.New("echo ⇒ renderer not registered")
)

// New creates an Echo instance.
func New() (e *Echo) {
	e = &Echo{
		uris: make(map[Handler]string),
	}
	e.router = NewRouter(e)
	e.pool.New = func() interface{} {
		return NewContext(nil, new(Response), e)
	}

	//----------
	// Defaults
	//----------

	e.HTTP2(true)
	e.SetMaxParam(5)
	e.notFoundHandler = func(c *Context) error {
		return NewHTTPError(http.StatusNotFound)
	}
	e.SetHTTPErrorHandler(func(err error, c *Context) {
		code := http.StatusInternalServerError
		msg := http.StatusText(code)
		if he, ok := err.(*HTTPError); ok {
			code = he.code
			msg = he.message
		}
		if e.Debug() {
			msg = err.Error()
		}
		http.Error(c.response, msg, code)
	})
	e.SetBinder(func(r *http.Request, v interface{}) error {
		ct := r.Header.Get(ContentType)
		err := UnsupportedMediaType
		if strings.HasPrefix(ct, ApplicationJSON) {
			err = json.NewDecoder(r.Body).Decode(v)
		} else if strings.HasPrefix(ct, ApplicationForm) {
			err = nil
		}
		return err
	})
	return
}

// Group creates a new sub router with prefix. It inherits all properties from
// the parent. Passing middleware overrides parent middleware.
func (e *Echo) Group(pfx string, m ...Middleware) *Echo {
	g := *e
	g.prefix = g.prefix + pfx
	if len(m) > 0 {
		g.middleware = nil
		g.Use(m...)
	}
	return &g
}

// Router returns router.
func (e *Echo) Router() *Router {
	return e.router
}

// HTTP2 enables HTTP2 support.
func (e *Echo) HTTP2(on bool) {
	e.http2 = on
}

// SetMaxParam sets the maximum number of path parameters allowed for the application.
// Default value is 5, good enough for many use cases.
func (e *Echo) SetMaxParam(n uint8) {
	e.maxParam = n
}

// SetHTTPErrorHandler registers an Echo.HTTPErrorHandler.
func (e *Echo) SetHTTPErrorHandler(h HTTPErrorHandler) {
	e.httpErrorHandler = h
}

// SetBinder registers a custom binder. It's invoked by Context.Bind().
func (e *Echo) SetBinder(b BindFunc) {
	e.binder = b
}

// SetRenderer registers an HTML template renderer. It's invoked by Context.Render().
func (e *Echo) SetRenderer(r Renderer) {
	e.renderer = r
}

// SetDebug sets debug mode.
func (e *Echo) SetDebug(on bool) {
	e.debug = on
}

// Debug returns debug mode.
func (e *Echo) Debug() bool {
	return e.debug
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
		wss.ServeHTTP(c.response.writer, c.request)
		return err
	})
}

func (e *Echo) add(method, path string, h Handler) {
	key := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	e.uris[key] = path
	e.router.Add(method, e.prefix+path, wrapHandler(h), e)
}

// Index serves index file.
func (e *Echo) Index(file string) {
	e.ServeFile("/", file)
}

// Favicon serves the default favicon - GET /favicon.ico.
func (e *Echo) Favicon(file string) {
	e.ServeFile("/favicon.ico", file)
}

// Static serves static files.
func (e *Echo) Static(path, root string) {
	fs := http.StripPrefix(path, http.FileServer(http.Dir(root)))
	e.Get(path+"*", func(c *Context) error {
		fs.ServeHTTP(c.response, c.request)
		return nil
	})
}

// ServeFile serves a file.
func (e *Echo) ServeFile(path, file string) {
	e.Get(path, func(c *Context) error {
		http.ServeFile(c.response, c.request, file)
		return nil
	})
}

// URI generates a URI from handler.
func (e *Echo) URI(h Handler, params ...interface{}) string {
	uri := new(bytes.Buffer)
	lp := len(params)
	n := 0
	key := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	if path, ok := e.uris[key]; ok {
		for i, l := 0, len(path); i < l; i++ {
			if path[i] == ':' && n < lp {
				for ; i < l && path[i] != '/'; i++ {
				}
				uri.WriteString(fmt.Sprintf("%v", params[n]))
				n++
			}
			if i < l {
				uri.WriteByte(path[i])
			}
		}
	}
	return uri.String()
}

// URL is an alias for URI
func (e *Echo) URL(h Handler, params ...interface{}) string {
	return e.URI(h, params...)
}

func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := e.pool.Get().(*Context)
	h, echo := e.router.Find(r.Method, r.URL.Path, c)
	if echo != nil {
		e = echo
	}
	c.reset(w, r, e)
	if h == nil {
		h = e.notFoundHandler
	}

	// Chain middleware with handler in the end
	for i := len(e.middleware) - 1; i >= 0; i-- {
		h = e.middleware[i](h)
	}

	// Execute chain
	if he := h(c); he != nil {
		e.httpErrorHandler(he, c)
	}

	e.pool.Put(c)
}

// Run runs a server.
func (e *Echo) Run(addr string) {
	s := &http.Server{Addr: addr}
	e.run(s)
}

// RunTLS runs a server with TLS configuration.
func (e *Echo) RunTLS(addr, certFile, keyFile string) {
	s := &http.Server{Addr: addr}
	e.run(s, certFile, keyFile)
}

// RunServer runs a custom server.
func (e *Echo) RunServer(srv *http.Server) {
	e.run(srv)
}

// RunTLSServer runs a custom server with TLS configuration.
func (e *Echo) RunTLSServer(srv *http.Server, certFile, keyFile string) {
	e.run(srv, certFile, keyFile)
}

func (e *Echo) run(s *http.Server, files ...string) {
	s.Handler = e
	if e.http2 {
		http2.ConfigureServer(s, nil)
	}
	if len(files) == 0 {
		log.Fatal(s.ListenAndServe())
	} else if len(files) == 2 {
		log.Fatal(s.ListenAndServeTLS(files[0], files[1]))
	} else {
		log.Fatal("echo => invalid TLS configuration")
	}
}

func NewHTTPError(code int, msg ...string) *HTTPError {
	he := &HTTPError{code: code, message: http.StatusText(code)}
	for _, m := range msg {
		he.message = m
	}
	return he
}

// Code returns code.
func (e *HTTPError) Code() int {
	return e.code
}

// Error returns message.
func (e *HTTPError) Error() string {
	return e.message
}

// wraps middleware
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
	case http.HandlerFunc:
		return wrapHTTPHandlerFuncMW(m)
	case func(http.ResponseWriter, *http.Request):
		return wrapHTTPHandlerFuncMW(m)
	default:
		panic("echo => unknown middleware")
	}
}

// Wraps HandlerFunc middleware
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

// Wraps http.HandlerFunc middleware
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

// wraps handler
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
		panic("echo ⇒ unknown handler")
	}
}

func init() {
	log.SetOutput(colorable.NewColorableStdout())
}
