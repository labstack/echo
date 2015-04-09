package echo

import (
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
)

type (
	Echo struct {
		Router          *router
		prefix          string
		middleware      []MiddlewareFunc
		maxParam        byte
		notFoundHandler HandlerFunc
		renderer        Renderer
		pool            sync.Pool
	}
	Middleware     interface{}
	MiddlewareFunc func(HandlerFunc) HandlerFunc
	Handler        interface{}
	HandlerFunc    func(*Context)
	Renderer       interface {
		Render(io.Writer, string, interface{}) error
	}
)

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

	// Errors
	ErrUnsupportedMediaType = errors.New("echo: unsupported media type")
)

// New creates an Echo instance.
func New() (e *Echo) {
	e = &Echo{
		maxParam: 5,
		notFoundHandler: func(c *Context) {
			http.Error(c.Response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		},
	}
	e.Router = NewRouter(e)
	e.pool.New = func() interface{} {
		return &Context{
			Response: &response{},
			params:   make(Params, e.maxParam),
			store:    make(store),
			echo:     e, // TODO: Do we need this?
		}
	}
	return
}

// NOP
func (h HandlerFunc) ServeHTTP(http.ResponseWriter, *http.Request) {
}

// Group creates a new sub router with prefix and inherits all properties from
// the parent. Passing middleware overrides parent middleware.
func (e *Echo) Group(pfx string, m ...Middleware) *Echo {
	g := *e
	g.prefix = pfx
	if len(m) > 0 {
		g.middleware = nil
		g.Use(m...)
	}
	return &g
}

// MaxParam sets the maximum allowed path parameters. Default is 5, good enough
// for many users.
func (e *Echo) MaxParam(n uint8) {
	e.maxParam = n
}

// NotFoundHandler sets a custom NotFound handler.
func (e *Echo) NotFoundHandler(h Handler) {
	e.notFoundHandler = wrapH(h)
}

// Renderer sets an HTML Renderer.
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

func (e *Echo) add(method, path string, h Handler) {
	e.Router.Add(method, e.prefix+path, wrapH(h), e)
}

// Static serves static files.
func (e *Echo) Static(path, root string) {
	fs := http.StripPrefix(path, http.FileServer(http.Dir(root)))
	e.Get(path+"/*", func(c *Context) {
		fs.ServeHTTP(c.Response, c.Request)
	})
}

// ServeFile serves a file.
func (e *Echo) ServeFile(path, file string) {
	e.Get(path, func(c *Context) {
		http.ServeFile(c.Response, c.Request, file)
	})
}

// Index serves index file.
func (e *Echo) Index(file string) {
	e.ServeFile("/", file)
}

func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := e.pool.Get().(*Context)
	h, echo := e.Router.Find(r.Method, r.URL.Path, c.params)
	if echo != nil {
		e = echo
	}
	if h == nil {
		h = e.notFoundHandler
	}
	c.reset(w, r, e)
	// Middleware
	for i := len(e.middleware) - 1; i >= 0; i-- {
		h = e.middleware[i](h)
	}
	// Handler
	h(c)
	e.pool.Put(c)
}

// Run a server
func (e *Echo) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, e))
}

// RunTLS a server
func (e *Echo) RunTLS(addr, certFile, keyFile string) {
	log.Fatal(http.ListenAndServeTLS(addr, certFile, keyFile, e))
}

// RunServer runs a custom server
func (e *Echo) RunServer(server *http.Server) {
	server.Handler = e
	log.Fatal(server.ListenAndServe())
}

// RunTLSServer runs a custom server with TLS configuration
func (e *Echo) RunTLSServer(server *http.Server, certFile, keyFile string) {
	server.Handler = e
	log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
}

// wraps Middleware
func wrapM(m Middleware) MiddlewareFunc {
	switch m := m.(type) {
	case func(*Context):
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) {
				m(c)
				h(c)
			}
		}
	case func(HandlerFunc) HandlerFunc:
		return MiddlewareFunc(m)
	case func(http.Handler) http.Handler:
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) {
				m(h).ServeHTTP(c.Response, c.Request)
				h(c)
			}
		}
	case http.Handler, http.HandlerFunc:
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) {
				m.(http.Handler).ServeHTTP(c.Response, c.Request)
				h(c)
			}
		}
	case func(http.ResponseWriter, *http.Request):
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) {
				m(c.Response, c.Request)
				h(c)
			}
		}
	default:
		panic("echo: unknown middleware")
	}
}

// wraps Handler
func wrapH(h Handler) HandlerFunc {
	switch h := h.(type) {
	case func(*Context):
		return HandlerFunc(h)
	case http.Handler, http.HandlerFunc:
		return func(c *Context) {
			h.(http.Handler).ServeHTTP(c.Response, c.Request)
		}
	case func(http.ResponseWriter, *http.Request):
		return func(c *Context) {
			h(c.Response, c.Request)
		}
	default:
		panic("echo: unknown handler")
	}
}
