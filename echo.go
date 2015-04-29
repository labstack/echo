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

	"github.com/labstack/gommon/color"
	"github.com/mattn/go-colorable"
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
	}
	Middleware     interface{}
	MiddlewareFunc func(HandlerFunc) HandlerFunc
	Handler        interface{}
	HandlerFunc    func(*Context) error

	// HTTPErrorHandler is a centralized HTTP error handler.
	HTTPErrorHandler func(int, error, *Context)

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

	MIMEJSON          = "application/json"
	MIMEText          = "text/plain"
	MIMEHTML          = "text/html"
	MIMEForm          = "application/x-www-form-urlencoded"
	MIMEMultipartForm = "multipart/form-data"

	HeaderAccept             = "Accept"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentLength      = "Content-Length"
	HeaderContentType        = "Content-Type"
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

	UnsupportedMediaType  = errors.New("echo: unsupported media type")
	RendererNotRegistered = errors.New("echo: renderer not registered")
)

// New creates an Echo instance.
func New() (e *Echo) {
	e = &Echo{
		uris: make(map[Handler]string),
	}
	e.Router = NewRouter(e)
	e.pool.New = func() interface{} {
		return &Context{
			Response: &response{},
			pnames:   make([]string, e.maxParam),
			pvalues:  make([]string, e.maxParam),
			store:    make(store),
		}
	}

	//----------
	// Defaults
	//----------

	e.MaxParam(5)
	e.NotFoundHandler(func(c *Context) {
		http.Error(c.Response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	e.HTTPErrorHandler(func(code int, err error, c *Context) {
		if err != nil {
			// TODO: Warning
			log.Printf("echo: %s", color.Yellow("http error handler not registered"))
			http.Error(c.Response, err.Error(), code)
		}
	})
	e.Binder(func(r *http.Request, v interface{}) error {
		ct := r.Header.Get(HeaderContentType)
		if strings.HasPrefix(ct, MIMEJSON) {
			return json.NewDecoder(r.Body).Decode(v)
		} else if strings.HasPrefix(ct, MIMEForm) {
			return nil
		}
		return UnsupportedMediaType
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

// MaxParam sets the maximum number of path parameters allowd for the application.
// Default value is 5, good enough for many use cases.
func (e *Echo) MaxParam(n uint8) {
	e.maxParam = n
}

// NotFoundHandler registers a custom NotFound handler used by router in case it
// doesn't find any registered handler for HTTP method and path.
func (e *Echo) NotFoundHandler(h Handler) {
	e.notFoundHandler = wrapH(h)
}

// HTTPErrorHandler registers an HTTP error handler.
func (e *Echo) HTTPErrorHandler(h HTTPErrorHandler) {
	e.httpErrorHandler = h
}

// Binder registers a custom binder. It's invoked by Context.Bind API.
func (e *Echo) Binder(b BindFunc) {
	e.binder = b
}

// Renderer registers an HTML template renderer. It's invoked by Context.Render
// API.
func (e *Echo) Renderer(r Renderer) {
	e.renderer = r
}

// Use adds handler to the middleware chain.
func (e *Echo) Use(m ...Middleware) {
	for _, h := range m {
		e.middleware = append(e.middleware, wrapM(h))
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

func (e *Echo) add(method, path string, h Handler) {
	key := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	e.uris[key] = path
	e.Router.Add(method, e.prefix+path, wrapH(h), e)
}

// Static serves static files.
func (e *Echo) Static(path, root string) {
	fs := http.StripPrefix(path, http.FileServer(http.Dir(root)))
	e.Get(path+"/*", func(c *Context) error {
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

// Index serves index file.
func (e *Echo) Index(file string) {
	e.ServeFile("/", file)
}

func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := e.pool.Get().(*Context)
	h, echo := e.Router.Find(r.Method, r.URL.Path, c)
	if echo != nil {
		e = echo
	}
	if h == nil {
		h = e.notFoundHandler
	}
	c.reset(w, r, e)

	// Chain middleware with handler in the end
	for i := len(e.middleware) - 1; i >= 0; i-- {
		h = e.middleware[i](h)
	}

	// Execute chain
	if err := h(c); err != nil {
		e.httpErrorHandler(http.StatusInternalServerError, err, c)
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

// wraps Middleware
func wrapM(m Middleware) MiddlewareFunc {
	switch m := m.(type) {
	case func(*Context):
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) error {
				m(c)
				return h(c)
			}
		}
	case func(*Context) error:
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) error {
				if err := m(c); err != nil {
					return err
				}
				return h(c)
			}
		}
	case func(HandlerFunc) HandlerFunc:
		return m
	case func(http.Handler) http.Handler:
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) (err error) {
				m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					c.Response.Writer = w
					c.Request = r
					err = h(c)
				})).ServeHTTP(c.Response.Writer, c.Request)
				return
			}
		}
	case http.Handler, http.HandlerFunc:
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) error {
				m.(http.Handler).ServeHTTP(c.Response.Writer, c.Request)
				return h(c)
			}
		}
	case func(http.ResponseWriter, *http.Request):
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) error {
				m(c.Response.Writer, c.Request)
				return h(c)
			}
		}
	case func(http.ResponseWriter, *http.Request) error:
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) error {
				if err := m(c.Response.Writer, c.Request); err != nil {
					return err
				}
				return h(c)
			}
		}
	default:
		panic("echo: unknown middleware")
	}
}

// wraps Handler
func wrapH(h Handler) HandlerFunc {
	switch h := h.(type) {
	case HandlerFunc:
		return h
	case func(*Context) error:
		return h
	case func(*Context):
		return func(c *Context) error {
			h(c)
			return nil
		}
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
	case func(http.ResponseWriter, *http.Request) error:
		return func(c *Context) error {
			return h(c.Response, c.Request)
		}
	default:
		panic("echo: unknown handler")
	}
}

func init() {
	log.SetOutput(colorable.NewColorableStdout())
}
