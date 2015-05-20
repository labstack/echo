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

	"github.com/mattn/go-colorable"
	"golang.org/x/net/websocket"
)

type (
	Echo struct {
		Router           *router
		prefix           string
		middleware       []MiddlewareFunc
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
		Code    int
		Message string
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

func NewHTTPError(code int, msgs ...string) *HTTPError {
	he := &HTTPError{Code: code}
	if len(msgs) == 0 {
		he.Message = http.StatusText(code)
	}
	return he
}

func (e *HTTPError) Error() string {
	return e.Message
}

// New creates an Echo instance.
func New() (e *Echo) {
	e = &Echo{
		uris: make(map[Handler]string),
	}
	e.Router = NewRouter(e)
	e.pool.New = func() interface{} {
		return NewContext(nil, new(Response), e)
	}

	//----------
	// Defaults
	//----------

	e.SetMaxParam(5)
	e.notFoundHandler = func(c *Context) error {
		return NewHTTPError(http.StatusNotFound)
	}
	e.SetHTTPErrorHandler(func(err error, c *Context) {
		code := http.StatusInternalServerError
		msg := http.StatusText(code)
		if he, ok := err.(*HTTPError); ok {
			code = he.Code
			msg = he.Message
		}
		if e.Debug() {
			msg = err.Error()
		}
		http.Error(c.Response, msg, code)
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
	e.Get(path, func(c *Context) *HTTPError {
		wss := websocket.Server{
			Handler: func(ws *websocket.Conn) {
				c.Socket = ws
				c.Response.status = http.StatusSwitchingProtocols
				h(c)
			},
		}
		wss.ServeHTTP(c.Response.writer, c.Request)
		return nil
	})
}

func (e *Echo) add(method, path string, h Handler) {
	key := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	e.uris[key] = path
	e.Router.Add(method, e.prefix+path, wrapHandler(h), e)
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
		fs.ServeHTTP(c.Response, c.Request)
		return nil
	})
}

// ServeFile serves a file.
func (e *Echo) ServeFile(path, file string) {
	e.Get(path, func(c *Context) error {
		http.ServeFile(c.Response, c.Request, file)
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
	h, echo := e.Router.Find(r.Method, r.URL.Path, c)
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
	log.Fatal(http.ListenAndServe(addr, e))
}

// RunTLS runs a server with TLS configuration.
func (e *Echo) RunTLS(addr, certFile, keyFile string) {
	log.Fatal(http.ListenAndServeTLS(addr, certFile, keyFile, e))
}

// RunServer runs a custom server.
func (e *Echo) RunServer(server *http.Server) {
	server.Handler = e
	log.Fatal(server.ListenAndServe())
}

// RunTLSServer runs a custom server with TLS configuration.
func (e *Echo) RunTLSServer(server *http.Server, certFile, keyFile string) {
	server.Handler = e
	log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
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
					c.Response.writer = w
					c.Request = r
					err = h(c)
				})).ServeHTTP(c.Response.writer, c.Request)
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
			if !c.Response.committed {
				m.ServeHTTP(c.Response.writer, c.Request)
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
			h.(http.Handler).ServeHTTP(c.Response, c.Request)
			return nil
		}
	case func(http.ResponseWriter, *http.Request):
		return func(c *Context) error {
			h(c.Response, c.Request)
			return nil
		}
	default:
		panic("echo ⇒ unknown handler")
	}
}

func init() {
	log.SetOutput(colorable.NewColorableStdout())
}
