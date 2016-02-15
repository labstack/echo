package echo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"encoding/xml"

	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/logger"
	"github.com/labstack/gommon/log"
)

type (
	Echo struct {
		prefix           string
		middleware       []Middleware
		http2            bool
		maxParam         *int
		notFoundHandler  HandlerFunc
		httpErrorHandler HTTPErrorHandler
		binder           Binder
		renderer         Renderer
		pool             sync.Pool
		debug            bool
		router           *Router
		logger           logger.Logger
	}

	Route struct {
		Method  string
		Path    string
		Handler string
	}

	HTTPError struct {
		code    int
		message string
	}

	Middleware interface {
		Handle(Handler) Handler
		Priority() int
	}

	MiddlewareFunc func(Handler) Handler

	Handler interface {
		Handle(Context) error
	}

	HandlerFunc func(Context) error

	// HTTPErrorHandler is a centralized HTTP error handler.
	HTTPErrorHandler func(error, Context)

	// Binder is the interface that wraps the Bind method.
	Binder interface {
		Bind(engine.Request, interface{}) error
	}

	binder struct {
	}

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
	OctetStream                      = "application/octet-stream"

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
	ErrNotFound              = NewHTTPError(http.StatusNotFound)
	ErrRendererNotRegistered = errors.New("renderer not registered")
	ErrInvalidRedirectCode   = errors.New("invalid redirect status code")

	//----------------
	// Error handlers
	//----------------

	notFoundHandler = HandlerFunc(func(c Context) error {
		return NewHTTPError(http.StatusNotFound)
	})

	methodNotAllowedHandler = HandlerFunc(func(c Context) error {
		return NewHTTPError(http.StatusMethodNotAllowed)
	})
)

// New creates an instance of Echo.
func New() (e *Echo) {
	e = &Echo{maxParam: new(int)}
	e.pool.New = func() interface{} {
		// NOTE: v2
		return NewContext(nil, nil, e)
	}
	e.router = NewRouter(e)
	e.middleware = []Middleware{e.router}

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

func (f MiddlewareFunc) Handle(h Handler) Handler {
	return f(h)
}

func (f MiddlewareFunc) Priority() int {
	return 1
}

func (f HandlerFunc) Handle(c Context) error {
	return f(c)
}

// Router returns router.
func (e *Echo) Router() *Router {
	return e.router
}

// SetLogger sets the logger instance.
func (e *Echo) SetLogger(l logger.Logger) {
	e.logger = l
}

// Logger returns the logger instance.
func (e *Echo) Logger() logger.Logger {
	return e.logger
}

// HTTP2 enable/disable HTTP2 support.
func (e *Echo) HTTP2(on bool) {
	e.http2 = on
}

// DefaultHTTPErrorHandler invokes the default HTTP error handler.
func (e *Echo) DefaultHTTPErrorHandler(err error, c Context) {
	code := http.StatusInternalServerError
	msg := http.StatusText(code)
	if he, ok := err.(*HTTPError); ok {
		code = he.code
		msg = he.message
	}
	if e.debug {
		msg = err.Error()
	}
	if !c.Response().Committed() {
		c.String(code, msg)
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

// Use adds handler to the middleware chain.
func (e *Echo) Use(middleware ...interface{}) {
	for _, m := range middleware {
		e.middleware = append(e.middleware, wrapMiddleware(m))
	}
}

// Connect adds a CONNECT route > handler to the router.
func (e *Echo) Connect(path string, handler interface{}, middleware ...interface{}) {
	e.add(CONNECT, path, handler, middleware...)
}

// Delete adds a DELETE route > handler to the router.
func (e *Echo) Delete(path string, handler interface{}, middleware ...interface{}) {
	e.add(DELETE, path, handler, middleware...)
}

// Get adds a GET route > handler to the router.
func (e *Echo) Get(path string, handler interface{}, middleware ...interface{}) {
	e.add(GET, path, handler, middleware...)
}

// Head adds a HEAD route > handler to the router.
func (e *Echo) Head(path string, handler interface{}, middleware ...interface{}) {
	e.add(HEAD, path, handler, middleware...)
}

// Options adds an OPTIONS route > handler to the router.
func (e *Echo) Options(path string, handler interface{}, middleware ...interface{}) {
	e.add(OPTIONS, path, handler, middleware...)
}

// Patch adds a PATCH route > handler to the router.
func (e *Echo) Patch(path string, handler interface{}, middleware ...interface{}) {
	e.add(PATCH, path, handler, middleware...)
}

// Post adds a POST route > handler to the router.
func (e *Echo) Post(path string, handler interface{}, middleware ...interface{}) {
	e.add(POST, path, handler, middleware...)
}

// Put adds a PUT route > handler to the router.
func (e *Echo) Put(path string, handler interface{}, middleware ...interface{}) {
	e.add(PUT, path, handler, middleware...)
}

// Trace adds a TRACE route > handler to the router.
func (e *Echo) Trace(path string, handler interface{}, middleware ...interface{}) {
	e.add(TRACE, path, handler, middleware...)
}

// Any adds a route > handler to the router for all HTTP methods.
func (e *Echo) Any(path string, handler interface{}, middleware ...interface{}) {
	for _, m := range methods {
		e.add(m, path, handler, middleware...)
	}
}

// Match adds a route > handler to the router for multiple HTTP methods provided.
func (e *Echo) Match(methods []string, path string, handler interface{}, middleware ...interface{}) {
	for _, m := range methods {
		e.add(m, path, handler, middleware...)
	}
}

// NOTE: v2
func (e *Echo) add(method, path string, handler interface{}, middleware ...interface{}) {
	h := wrapHandler(handler)
	name := handlerName(handler)
	e.router.Add(method, path, HandlerFunc(func(c Context) error {
		for _, m := range middleware {
			h = wrapMiddleware(m).Handle(h)
		}
		return h.Handle(c)
	}), e)
	r := Route{
		Method:  method,
		Path:    path,
		Handler: name,
	}
	e.router.routes = append(e.router.routes, r)
}

// Group creates a new sub-router with prefix.
func (e *Echo) Group(prefix string, middleware ...interface{}) (g *Group) {
	g = &Group{prefix: prefix, echo: e}
	g.Use(middleware...)
	return
}

// URI generates a URI from handler.
func (e *Echo) URI(handler interface{}, params ...interface{}) string {
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
func (e *Echo) URL(handler interface{}, params ...interface{}) string {
	return e.URI(handler, params...)
}

// Routes returns the registered routes.
func (e *Echo) Routes() []Route {
	return e.router.routes
}

func (e *Echo) ServeHTTP(req engine.Request, res engine.Response) {
	c := e.pool.Get().(*context)
	c.reset(req, res)
	h := Handler(c)

	// Chain middleware with handler in the end
	for i := len(e.middleware) - 1; i >= 0; i-- {
		h = e.middleware[i].Handle(h)
	}

	// Execute chain
	if err := h.Handle(c); err != nil {
		e.httpErrorHandler(err, c)
	}

	e.pool.Put(c)
}

// Run starts the HTTP engine.
func (e *Echo) Run(eng engine.Engine) {
	eng.SetHandler(e.ServeHTTP)
	eng.SetLogger(e.logger)
	eng.Start()
}

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

func (binder) Bind(r engine.Request, i interface{}) (err error) {
	ct := r.Header().Get(ContentType)
	err = ErrUnsupportedMediaType
	if strings.HasPrefix(ct, ApplicationJSON) {
		if err = json.NewDecoder(r.Body()).Decode(i); err != nil {
			err = NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else if strings.HasPrefix(ct, ApplicationXML) {
		if err = xml.NewDecoder(r.Body()).Decode(i); err != nil {
			err = NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}
	return
}

func wrapMiddleware(m interface{}) Middleware {
	switch m := m.(type) {
	case Middleware:
		return m
	case MiddlewareFunc:
		return m
	case func(Handler) Handler:
		return MiddlewareFunc(m)
	default:
		panic("invalid middleware")
	}
}

func wrapHandler(h interface{}) Handler {
	switch h := h.(type) {
	case Handler:
		return h
	case HandlerFunc:
		return h
	case func(Context) error:
		return HandlerFunc(h)
	default:
		panic("echo => invalid handler")
	}
}

func handlerName(h interface{}) string {
	switch h := h.(type) {
	case Handler:
		t := reflect.TypeOf(h)
		return fmt.Sprintf("%s » %s", t.PkgPath(), t.Name())
	case HandlerFunc, func(Context) error:
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	default:
		panic("echo => invalid handler")
	}
}
