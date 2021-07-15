/*
Package echo implements high performance, minimalist Go web framework.

Example:

  package main

  import (
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
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
    if err := e.Start(":8080"); err != http.ErrServerClosed {
		  log.Fatal(err)
	  }
  }

Learn more at https://echo.labstack.com
*/
package echo

import (
	stdContext "context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
)

// Echo is the top-level framework instance.
// Note: replacing/nilling public fields is not coroutine/thread-safe and can cause data-races/panics.
type Echo struct {
	// premiddleware are middlewares that are run for every request before routing is done
	premiddleware []MiddlewareFunc
	// middleware are middlewares that are run after router found a matching route (not found and method not found are also matches)
	middleware []MiddlewareFunc

	router        Router
	routers       map[string]Router
	routerCreator func(e *Echo) Router

	contextPool               sync.Pool
	contextPathParamAllocSize int

	// NewContextFunc allows using custom context implementations, instead of default *echo.context
	NewContextFunc   func(pathParamAllocSize int) EditableContext
	Debug            bool
	HTTPErrorHandler HTTPErrorHandler
	Binder           Binder
	JSONSerializer   JSONSerializer
	Validator        Validator
	Renderer         Renderer
	Logger           Logger
	IPExtractor      IPExtractor
	// Filesystem is file system used by Static and File handler to access files.
	// Defaults to os.DirFS(".")
	Filesystem fs.FS
}

// JSONSerializer is the interface that encodes and decodes JSON to and from interfaces.
type JSONSerializer interface {
	Serialize(c Context, i interface{}, indent string) error
	Deserialize(c Context, i interface{}) error
}

// HTTPErrorHandler is a centralized HTTP error handler.
type HTTPErrorHandler func(c Context, err error)

// HandlerFunc defines a function to serve HTTP requests.
type HandlerFunc func(c Context) error

// MiddlewareFunc defines a function to process middleware.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// MiddlewareConfigurator defines interface for creating middleware handlers with possibility to return configuration errors instead of panicking.
type MiddlewareConfigurator interface {
	ToMiddleware() (MiddlewareFunc, error)
}

// Validator is the interface that wraps the Validate function.
type Validator interface {
	Validate(i interface{}) error
}

// Renderer is the interface that wraps the Render function.
type Renderer interface {
	Render(io.Writer, string, interface{}, Context) error
}

// Map defines a generic map of type `map[string]interface{}`.
type Map map[string]interface{}

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextXML                          = "text/xml"
	MIMETextXMLCharsetUTF8               = MIMETextXML + "; " + charsetUTF8
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
	charsetUTF8 = "charset=UTF-8"
	// PROPFIND Method can be used on collection and property resources.
	PROPFIND = "PROPFIND"
	// REPORT Method can be used to get information about a resource, see rfc 3253
	REPORT = "REPORT"
)

// Headers
const (
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXXSSProtection                  = "X-XSS-Protection"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderXCSRFToken                      = "X-CSRF-Token"
	HeaderReferrerPolicy                  = "Referrer-Policy"
)

const (
	// Version of Echo
	Version = "5.0.X"
)

var methods = [...]string{
	http.MethodConnect,
	http.MethodDelete,
	http.MethodGet,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPatch,
	http.MethodPost,
	PROPFIND,
	http.MethodPut,
	http.MethodTrace,
	REPORT,
}

// New creates an instance of Echo.
func New() *Echo {
	logger := newJSONLogger(os.Stdout)
	e := &Echo{
		Logger:         logger,
		Filesystem:     os.DirFS("."),
		Binder:         &DefaultBinder{},
		JSONSerializer: &DefaultJSONSerializer{},

		routers: make(map[string]Router),
		routerCreator: func(ec *Echo) Router {
			return NewRouter(ec, RouterConfig{})
		},
	}

	e.router = NewRouter(e, RouterConfig{})
	e.HTTPErrorHandler = DefaultHTTPErrorHandler(false)
	e.contextPool.New = func() interface{} {
		if e.NewContextFunc != nil {
			return e.NewContextFunc(e.contextPathParamAllocSize)
		}
		return e.NewContext(nil, nil)
	}
	return e
}

// NewContext returns a Context instance.
func (e *Echo) NewContext(r *http.Request, w http.ResponseWriter) Context {
	p := make(PathParams, e.contextPathParamAllocSize)
	return &context{
		request:    r,
		response:   NewResponse(w, e),
		store:      make(Map),
		echo:       e,
		pathParams: &p,
		matchType:  RouteMatchUnknown,
		route:      nil,
		path:       "",
	}
}

// Router returns the default router.
func (e *Echo) Router() Router {
	return e.router
}

// Routers returns the map of host => router.
func (e *Echo) Routers() map[string]Router {
	return e.routers
}

// RouterFor returns Router for given host.
func (e *Echo) RouterFor(host string) Router {
	return e.routers[host]
}

// ResetRouterCreator resets callback for creating new router instances.
// Note: current (default) router is immediately replaced with router created with creator func and vhost routers are cleared.
func (e *Echo) ResetRouterCreator(creator func(e *Echo) Router) {
	e.routerCreator = creator
	e.router = creator(e)
	e.routers = make(map[string]Router)
}

// DefaultHTTPErrorHandler creates new default HTTP error handler implementation. It sends a JSON response
// with status code. `exposeError` parameter decides if returned message will contain also error message or not
//
// Note: DefaultHTTPErrorHandler does not log errors. Use middleware for it if errors need to be logged (separately)
// Note: In case errors happens in middleware call-chain that is returning from handler (which did not return an error).
// When handler has already sent response (ala c.JSON()) and there is error in middleware that is returning from
// handler. Then the error that global error handler received will be ignored because we have already "commited" the
// response and status code header has been sent to the client.
func DefaultHTTPErrorHandler(exposeError bool) HTTPErrorHandler {
	return func(c Context, err error) {
		if c.Response().Committed {
			return
		}

		he := &HTTPError{
			Code:    http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
		if errors.As(err, &he) {
			if he.Internal != nil { // max 2 levels of checks even if internal could have also internal
				errors.As(he.Internal, &he)
			}
		}

		// Issue #1426
		code := he.Code
		message := he.Message
		if m, ok := he.Message.(string); ok {
			if exposeError {
				message = Map{"message": m, "error": err.Error()}
			} else {
				message = Map{"message": m}
			}
		}

		// Send response
		var cErr error
		if c.Request().Method == http.MethodHead { // Issue #608
			cErr = c.NoContent(he.Code)
		} else {
			cErr = c.JSON(code, message)
		}
		if cErr != nil {
			c.Echo().Logger.Error(err) // truly rare case. ala client already disconnected
		}
	}
}

// Pre adds middleware to the chain which is run before router tries to find matching route.
// Meaning middleware is executed even for 404 (not found) cases.
func (e *Echo) Pre(middleware ...MiddlewareFunc) {
	e.premiddleware = append(e.premiddleware, middleware...)
}

// Use adds middleware to the chain which is run after router has found matching route and before route/request handler method is executed.
func (e *Echo) Use(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the
// router with optional route-level middleware. Panics on error.
func (e *Echo) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return e.Add(http.MethodConnect, path, h, m...)
}

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware. Panics on error.
func (e *Echo) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return e.Add(http.MethodDelete, path, h, m...)
}

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware. Panics on error.
func (e *Echo) GET(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return e.Add(http.MethodGet, path, h, m...)
}

// HEAD registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware. Panics on error.
func (e *Echo) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return e.Add(http.MethodHead, path, h, m...)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware. Panics on error.
func (e *Echo) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return e.Add(http.MethodOptions, path, h, m...)
}

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware. Panics on error.
func (e *Echo) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return e.Add(http.MethodPatch, path, h, m...)
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware. Panics on error.
func (e *Echo) POST(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return e.Add(http.MethodPost, path, h, m...)
}

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware. Panics on error.
func (e *Echo) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return e.Add(http.MethodPut, path, h, m...)
}

// TRACE registers a new TRACE route for a path with matching handler in the
// router with optional route-level middleware. Panics on error.
func (e *Echo) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo {
	return e.Add(http.MethodTrace, path, h, m...)
}

// Any registers a new route for all supported HTTP methods and path with matching handler
// in the router with optional route-level middleware. Panics on error.
func (e *Echo) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) Routes {
	errs := make([]error, 0)
	ris := make(Routes, 0)
	for _, m := range methods {
		ri, err := e.AddRoute(Route{
			Method:      m,
			Path:        path,
			Handler:     handler,
			Middlewares: middleware,
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ris = append(ris, ri)
	}
	if len(errs) > 0 {
		panic(errs) // this is how `v4` handles errors. `v5` has methods to have panic-free usage
	}
	return ris
}

// Match registers a new route for multiple HTTP methods and path with matching
// handler in the router with optional route-level middleware. Panics on error.
func (e *Echo) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) Routes {
	errs := make([]error, 0)
	ris := make(Routes, 0)
	for _, m := range methods {
		ri, err := e.AddRoute(Route{
			Method:      m,
			Path:        path,
			Handler:     handler,
			Middlewares: middleware,
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ris = append(ris, ri)
	}
	if len(errs) > 0 {
		panic(errs) // this is how `v4` handles errors. `v5` has methods to have panic-free usage
	}
	return ris
}

// Static registers a new route with path prefix to serve static files from the provided root directory. Panics on error.
func (e *Echo) Static(prefix, root string, middleware ...MiddlewareFunc) RouteInfo {
	return e.Add(
		http.MethodGet,
		prefix+"*",
		StaticDirectoryHandler(root, false),
		middleware...,
	)
}

// StaticDirectoryHandler creates handler function to serve files from given root path
func StaticDirectoryHandler(root string, disablePathUnescaping bool) HandlerFunc {
	if root == "" {
		root = "." // For security we want to restrict to CWD.
	}
	return func(c Context) error {
		p := c.PathParam("*")
		if !disablePathUnescaping { // when router is already unescaping we do not want to do is twice
			tmpPath, err := url.PathUnescape(p)
			if err != nil {
				return fmt.Errorf("failed to unescape path variable: %w", err)
			}
			p = tmpPath
		}

		name := filepath.Join(root, filepath.Clean("/"+p)) // "/"+ for security
		fi, err := fs.Stat(c.Echo().Filesystem, name)
		if err != nil {
			// The access path does not exist
			return ErrNotFound
		}

		// If the request is for a directory and does not end with "/"
		p = c.Request().URL.Path // path must not be empty.
		if fi.IsDir() && p[len(p)-1] != '/' {
			// Redirect to ends with "/"
			return c.Redirect(http.StatusMovedPermanently, p+"/")
		}
		return c.File(name)
	}
}

// File registers a new route with path to serve a static file with optional route-level middleware. Panics on error.
func (e *Echo) File(path, file string, middleware ...MiddlewareFunc) RouteInfo {
	handler := func(c Context) error {
		return c.File(file)
	}
	return e.Add(http.MethodGet, path, handler, middleware...)
}

// AddRoute registers a new Route with default host Router
func (e *Echo) AddRoute(route Routable) (RouteInfo, error) {
	return e.add("", route)
}

func (e *Echo) add(host string, route Routable) (RouteInfo, error) {
	router := e.findRouter(host)
	ri, err := router.Add(route)
	if err != nil {
		return nil, err
	}

	paramsCount := len(ri.Params())
	if paramsCount > e.contextPathParamAllocSize {
		e.contextPathParamAllocSize = paramsCount
	}
	return ri, nil
}

// Add registers a new route for an HTTP method and path with matching handler
// in the router with optional route-level middleware.
func (e *Echo) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouteInfo {
	ri, err := e.add(
		"",
		Route{
			Method:      method,
			Path:        path,
			Handler:     handler,
			Middlewares: middleware,
			Name:        "",
		},
	)
	if err != nil {
		panic(err) // this is how `v4` handles errors. `v5` has methods to have panic-free usage
	}
	return ri
}

// Host creates a new router group for the provided host and optional host-level middleware.
func (e *Echo) Host(name string, m ...MiddlewareFunc) (g *Group) {
	e.routers[name] = e.routerCreator(e)
	g = &Group{host: name, echo: e}
	g.Use(m...)
	return
}

// Group creates a new router group with prefix and optional group-level middleware.
func (e *Echo) Group(prefix string, m ...MiddlewareFunc) (g *Group) {
	g = &Group{prefix: prefix, echo: e}
	g.Use(m...)
	return
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must return the context by calling `ReleaseContext()`.
func (e *Echo) AcquireContext() Context {
	return e.contextPool.Get().(Context)
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (e *Echo) ReleaseContext(c Context) {
	e.contextPool.Put(c)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var c EditableContext
	if e.NewContextFunc != nil {
		// NOTE: we are not casting always context to EditableContext because casting to interface vs pointer to struct is
		// "significantly" slower. Echo Context interface has way to many methods so these checks take time.
		// These are benchmarks with 1.16:
		// * interface extending another interface = +24% slower (3233 ns/op vs 2605 ns/op)
		// * interface (not extending any, just methods)= +14% slower
		//
		// Quote from https://stackoverflow.com/a/31584377
		// "it's even worse with interface-to-interface assertion, because you also need to ensure that the type implements the interface."
		//
		// So most of the time we do not need custom context type and simple IF + cast to pointer to struct is fast enough.
		c = e.contextPool.Get().(EditableContext)
	} else {
		c = e.contextPool.Get().(*context)
	}
	c.Reset(r, w)
	var h func(c Context) error

	if e.premiddleware == nil {
		params := c.RawPathParams()
		match := e.findRouter(r.Host).Match(r, params)

		c.SetRawPathParams(params)
		c.SetPath(match.RoutePath)
		c.SetRouteInfo(match.RouteInfo)
		c.SetRouteMatchType(match.Type)
		h = applyMiddleware(match.Handler, e.middleware...)
	} else {
		h = func(cc Context) error {
			params := c.RawPathParams()
			match := e.findRouter(r.Host).Match(r, params)
			// NOTE: router will be executed after pre middlewares have been run. We assume here that context we receive after pre middlewares
			// is the same we began with. If not - this is use-case we do not support and is probably abuse from developer.
			c.SetRawPathParams(params)
			c.SetPath(match.RoutePath)
			c.SetRouteInfo(match.RouteInfo)
			c.SetRouteMatchType(match.Type)
			h1 := applyMiddleware(match.Handler, e.middleware...)
			return h1(cc)
		}
		h = applyMiddleware(h, e.premiddleware...)
	}

	// Execute chain
	if err := h(c); err != nil {
		e.HTTPErrorHandler(c, err)
	}

	e.contextPool.Put(c)
}

// Start stars HTTP server on given address with Echo as a handler serving requests. The server can be shutdown by
// sending os.Interrupt signal with `ctrl+c`.
//
// Note: this method is created for use in examples/demos and is deliberately simple without providing configuration
// options.
//
// In need of customization use:
// 	sc := echo.StartConfig{Address: ":8080"}
//	if err := sc.Start(e); err != http.ErrServerClosed {
//		log.Fatal(err)
//	}
// // or standard library `http.Server`
// 	s := http.Server{Addr: ":8080", Handler: e}
//	if err := s.ListenAndServe(); err != http.ErrServerClosed {
//		log.Fatal(err)
//	}
func (e *Echo) Start(address string) error {
	sc := StartConfig{Address: address}
	ctx, cancel := signal.NotifyContext(stdContext.Background(), os.Interrupt) // start shutdown process on ctrl+c
	defer cancel()
	sc.GracefulContext = ctx

	return sc.Start(e)
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
				c.SetRequest(r)
				c.SetResponse(NewResponse(w, c.Echo()))
				err = next(c)
			})).ServeHTTP(c.Response(), c.Request())
			return
		}
	}
}

func (e *Echo) findRouter(host string) Router {
	if len(e.routers) > 0 {
		if r, ok := e.routers[host]; ok {
			return r
		}
	}
	return e.router
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
