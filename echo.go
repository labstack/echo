package echo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
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
		middleware       []MiddlewareFunc
		http2            bool
		maxParam         *int
		notFoundHandler  HandlerFunc
		httpErrorHandler HTTPErrorHandler
		binder           Binder
		renderer         Renderer
		pool             sync.Pool
		debug            bool
		hook             engine.HandlerFunc
		autoIndex        bool
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
		Process(HandlerFunc) HandlerFunc
	}

	MiddlewareFunc func(HandlerFunc) HandlerFunc

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

	UnsupportedMediaType  = NewHTTPError(http.StatusUnsupportedMediaType)
	RendererNotRegistered = errors.New("renderer not registered")
	InvalidRedirectCode   = errors.New("invalid redirect status code")

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

func (f MiddlewareFunc) Process(h HandlerFunc) HandlerFunc {
	return f(h)
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

// AutoIndex enable/disable automatically creating an index page for the directory.
func (e *Echo) AutoIndex(on bool) {
	e.autoIndex = on
}

// Hook registers a callback which is invoked from `Echo#ServerHTTP` as the first
// statement. Hook is useful if you want to modify response/response objects even
// before it hits the router or any middleware.
func (e *Echo) Hook(h engine.HandlerFunc) {
	e.hook = h
}

// Use adds handler to the middleware chain.
func (e *Echo) Use(middleware ...interface{}) {
	for _, m := range middleware {
		e.middleware = append(e.middleware, wrapMiddleware(m))
	}
}

// Connect adds a CONNECT route > handler to the router.
func (e *Echo) Connect(path string, handler interface{}) {
	e.add(CONNECT, path, handler)
}

// Delete adds a DELETE route > handler to the router.
func (e *Echo) Delete(path string, handler interface{}) {
	e.add(DELETE, path, handler)
}

// Get adds a GET route > handler to the router.
func (e *Echo) Get(path string, handler interface{}) {
	e.add(GET, path, handler)
}

// Head adds a HEAD route > handler to the router.
func (e *Echo) Head(path string, handler interface{}) {
	e.add(HEAD, path, handler)
}

// Options adds an OPTIONS route > handler to the router.
func (e *Echo) Options(path string, handler interface{}) {
	e.add(OPTIONS, path, handler)
}

// Patch adds a PATCH route > handler to the router.
func (e *Echo) Patch(path string, handler interface{}) {
	e.add(PATCH, path, handler)
}

// Post adds a POST route > handler to the router.
func (e *Echo) Post(path string, handler interface{}) {
	e.add(POST, path, handler)
}

// Put adds a PUT route > handler to the router.
func (e *Echo) Put(path string, handler interface{}) {
	e.add(PUT, path, handler)
}

// Trace adds a TRACE route > handler to the router.
func (e *Echo) Trace(path string, handler interface{}) {
	e.add(TRACE, path, handler)
}

// Any adds a route > handler to the router for all HTTP methods.
func (e *Echo) Any(path string, handler interface{}) {
	for _, m := range methods {
		e.add(m, path, handler)
	}
}

// Match adds a route > handler to the router for multiple HTTP methods provided.
func (e *Echo) Match(methods []string, path string, handler interface{}) {
	for _, m := range methods {
		e.add(m, path, handler)
	}
}

// NOTE: v2
func (e *Echo) add(method, path string, h interface{}) {
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
	e.Get(path+"*", func(c Context) error {
		return e.serveFile(dir, c.P(0), c) // Param `_*`
	})
}

// ServeFile serves a file.
func (e *Echo) ServeFile(path, file string) {
	e.Get(path, func(c Context) error {
		dir, file := filepath.Split(file)
		return e.serveFile(dir, file, c)
	})
}

func (e *Echo) serveFile(dir, file string, c Context) (err error) {
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
	c.Response().WriteHeader(http.StatusOK)
	io.Copy(c.Response(), f)
	// TODO:
	// http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
	return
}

func listDir(d http.File, c Context) (err error) {
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
func (e *Echo) Group(prefix string, m ...MiddlewareFunc) *Group {
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
func (e *Echo) URI(h HandlerFunc, params ...interface{}) string {
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
func (e *Echo) URL(h HandlerFunc, params ...interface{}) string {
	return e.URI(h, params...)
}

// Routes returns the registered routes.
func (e *Echo) Routes() []Route {
	return e.router.routes
}

func (e *Echo) ServeHTTP(req engine.Request, res engine.Response) {
	if e.hook != nil {
		e.hook(req, res)
	}

	c := e.pool.Get().(*context)
	h, e := e.router.Find(req.Method(), req.URL().Path(), c)
	c.reset(req, res, e)

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
// func (e *Echo) Server(addr string) *http.Server {
// 	s := &http.Server{Addr: addr, Handler: e}
// 	// TODO: Remove in Go 1.6+
// 	if e.http2 {
// 		http2.ConfigureServer(s, nil)
// 	}
// 	return s
// }

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
	err = UnsupportedMediaType
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

func wrapMiddleware(m interface{}) MiddlewareFunc {
	switch m := m.(type) {
	case Middleware:
		return m.Process
	case MiddlewareFunc:
		return m
	case func(HandlerFunc) HandlerFunc:
		return m
	default:
		panic("invalid middleware")
	}
}

func wrapHandler(h interface{}) HandlerFunc {
	switch h := h.(type) {
	case Handler:
		return h.Handle
	case HandlerFunc:
		return h
	case func(Context) error:
		return h
	default:
		panic("invalid handler")
	}
}
