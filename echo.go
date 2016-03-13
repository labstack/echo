package echo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"encoding/xml"

	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
)

type (
	Echo struct {
		prefix           string
		middleware       []Middleware
		head             Handler
		maxParam         *int
		notFoundHandler  HandlerFunc
		httpErrorHandler HTTPErrorHandler
		binder           Binder
		renderer         Renderer
		pool             sync.Pool
		debug            bool
		router           *Router
		logger           *log.Logger
	}

	Route struct {
		Method  string
		Path    string
		Handler string
	}

	HTTPError struct {
		Code    int
		Message string
	}

	Middleware interface {
		Handle(Handler) Handler
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
		Bind(interface{}, Context) error
	}

	binder struct {
	}

	// Validator is the interface that wraps the Validate method.
	Validator interface {
		Validate() error
	}

	// Renderer is the interface that wraps the Render method.
	Renderer interface {
		Render(io.Writer, string, interface{}, Context) error
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
	LastModified       = "Last-Modified"
	Location           = "Location"
	Upgrade            = "Upgrade"
	Vary               = "Vary"
	WWWAuthenticate    = "WWW-Authenticate"
	XForwardedFor      = "X-Forwarded-For"
	XRealIP            = "X-Real-IP"
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
	ErrUnauthorized          = NewHTTPError(http.StatusUnauthorized)
	ErrMethodNotAllowed      = NewHTTPError(http.StatusMethodNotAllowed)
	ErrRendererNotRegistered = errors.New("renderer not registered")
	ErrInvalidRedirectCode   = errors.New("invalid redirect status code")

	//----------------
	// Error handlers
	//----------------

	notFoundHandler = HandlerFunc(func(c Context) error {
		return ErrNotFound
	})

	methodNotAllowedHandler = HandlerFunc(func(c Context) error {
		return ErrMethodNotAllowed
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
	e.head = e.router.Handle(nil)

	//----------
	// Defaults
	//----------

	e.SetHTTPErrorHandler(e.DefaultHTTPErrorHandler)
	e.SetBinder(&binder{})

	// Logger
	e.logger = log.New("echo")
	e.logger.SetLevel(log.FATAL)

	return
}

func (m MiddlewareFunc) Handle(h Handler) Handler {
	return m(h)
}

func (h HandlerFunc) Handle(c Context) error {
	return h(c)
}

// Router returns router.
func (e *Echo) Router() *Router {
	return e.router
}

// SetLogPrefix sets the prefix for the logger. Default value is `echo`.
func (e *Echo) SetLogPrefix(prefix string) {
	e.logger.SetPrefix(prefix)
}

// SetLogOutput sets the output destination for the logger. Default value is `os.Std*`
func (e *Echo) SetLogOutput(w io.Writer) {
	e.logger.SetOutput(w)
}

// SetLogLevel sets the log level for the logger. Default value is `log.FATAL`.
func (e *Echo) SetLogLevel(l log.Level) {
	e.logger.SetLevel(l)
}

// Logger returns the logger instance.
func (e *Echo) Logger() *log.Logger {
	return e.logger
}

// DefaultHTTPErrorHandler invokes the default HTTP error handler.
func (e *Echo) DefaultHTTPErrorHandler(err error, c Context) {
	code := http.StatusInternalServerError
	msg := http.StatusText(code)
	if he, ok := err.(*HTTPError); ok {
		code = he.Code
		msg = he.Message
	}
	if e.debug {
		msg = err.Error()
	}
	if !c.Response().Committed() {
		c.String(code, msg)
	}
	e.logger.Debug(err)
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
	e.SetLogLevel(log.DEBUG)
}

// Debug returns debug mode (enabled or disabled).
func (e *Echo) Debug() bool {
	return e.debug
}

// Use adds handler to the middleware chain.
func (e *Echo) Use(middleware ...Middleware) {
	e.middleware = append(e.middleware, middleware...)
	m := append(e.middleware, e.router)

	// Chain middleware
	for i := len(m) - 1; i >= 0; i-- {
		e.head = m[i].Handle(e.head)
	}
}

// Connect adds a CONNECT route > handler to the router.
func (e *Echo) Connect(path string, h Handler, m ...Middleware) {
	e.add(CONNECT, path, h, m...)
}

// Delete adds a DELETE route > handler to the router.
func (e *Echo) Delete(path string, h Handler, m ...Middleware) {
	e.add(DELETE, path, h, m...)
}

// Get adds a GET route > handler to the router.
func (e *Echo) Get(path string, h Handler, m ...Middleware) {
	e.add(GET, path, h, m...)
}

// Head adds a HEAD route > handler to the router.
func (e *Echo) Head(path string, h Handler, m ...Middleware) {
	e.add(HEAD, path, h, m...)
}

// Options adds an OPTIONS route > handler to the router.
func (e *Echo) Options(path string, h Handler, m ...Middleware) {
	e.add(OPTIONS, path, h, m...)
}

// Patch adds a PATCH route > handler to the router.
func (e *Echo) Patch(path string, h Handler, m ...Middleware) {
	e.add(PATCH, path, h, m...)
}

// Post adds a POST route > handler to the router.
func (e *Echo) Post(path string, h Handler, m ...Middleware) {
	e.add(POST, path, h, m...)
}

// Put adds a PUT route > handler to the router.
func (e *Echo) Put(path string, h Handler, m ...Middleware) {
	e.add(PUT, path, h, m...)
}

// Trace adds a TRACE route > handler to the router.
func (e *Echo) Trace(path string, h Handler, m ...Middleware) {
	e.add(TRACE, path, h, m...)
}

// Any adds a route > handler to the router for all HTTP methods.
func (e *Echo) Any(path string, handler Handler, middleware ...Middleware) {
	for _, m := range methods {
		e.add(m, path, handler, middleware...)
	}
}

// Match adds a route > handler to the router for multiple HTTP methods provided.
func (e *Echo) Match(methods []string, path string, handler Handler, middleware ...Middleware) {
	for _, m := range methods {
		e.add(m, path, handler, middleware...)
	}
}

// Static serves files from provided `root` directory for `/<prefix>*` HTTP path.
func (e *Echo) Static(prefix, root string) {
	e.Get(prefix+"*", HandlerFunc(func(c Context) error {
		return c.File(path.Join(root, c.P(0))) // Param `_`
	}))
}

// File serves provided file for `/<path>` HTTP path.
func (e *Echo) File(path, file string) {
	e.Get(path, HandlerFunc(func(c Context) error {
		return c.File(file)
	}))
}

func (e *Echo) add(method, path string, handler Handler, middleware ...Middleware) {
	name := handlerName(handler)
	e.router.Add(method, path, HandlerFunc(func(c Context) error {
		for _, m := range middleware {
			handler = m.Handle(handler)
		}
		return handler.Handle(c)
	}), e)
	r := Route{
		Method:  method,
		Path:    path,
		Handler: name,
	}
	e.router.routes = append(e.router.routes, r)
}

// Group creates a new sub-router with prefix.
func (e *Echo) Group(prefix string, m ...Middleware) (g *Group) {
	g = &Group{prefix: prefix, echo: e}
	g.Use(m...)
	return
}

// URI generates a URI from handler.
func (e *Echo) URI(handler Handler, params ...interface{}) string {
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
func (e *Echo) URL(h Handler, params ...interface{}) string {
	return e.URI(h, params...)
}

// Routes returns the registered routes.
func (e *Echo) Routes() []Route {
	return e.router.routes
}

func (e *Echo) ServeHTTP(req engine.Request, res engine.Response) {
	c := e.pool.Get().(*context)
	c.reset(req, res)

	// Execute chain
	if err := e.head.Handle(c); err != nil {
		e.httpErrorHandler(err, c)
	}

	e.pool.Put(c)
}

// Run starts the HTTP engine.
func (e *Echo) Run(eng engine.Engine) {
	eng.SetHandler(e)
	eng.SetLogger(e.logger)
	eng.Start()
}

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

func (binder) Bind(i interface{}, c Context) (err error) {
	req := c.Request()
	ct := req.Header().Get(ContentType)
	err = ErrUnsupportedMediaType
	if strings.HasPrefix(ct, ApplicationJSON) {
		if err = json.NewDecoder(req.Body()).Decode(i); err != nil {
			err = NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else if strings.HasPrefix(ct, ApplicationXML) {
		if err = xml.NewDecoder(req.Body()).Decode(i); err != nil {
			err = NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}
	return
}

// WrapMiddleware wrap `echo.Handler` into `echo.MiddlewareFunc`.
func WrapMiddleware(h Handler) MiddlewareFunc {
	return func(next Handler) Handler {
		return HandlerFunc(func(c Context) error {
			if err := h.Handle(c); err != nil {
				return err
			}
			return next.Handle(c)
		})
	}
}

func handlerName(h Handler) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}
