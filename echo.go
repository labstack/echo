package echo

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
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
		pool             sync.Pool
	}
	Middleware     interface{}
	MiddlewareFunc func(HandlerFunc) (HandlerFunc, error)
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
	ErrNoRenderer           = errors.New("echo: renderer not registered")
)

// New creates an Echo instance.
func New() (e *Echo) {
	e = &Echo{
		maxParam: 5,
		notFoundHandler: func(c *Context) error {
			http.Error(c.Response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return nil
		},
		httpErrorHandler: func(err error, c *Context) {
			http.Error(c.Response, err.Error(), http.StatusInternalServerError)
		},
		binder: func(r *http.Request, v interface{}) error {
			ct := r.Header.Get(HeaderContentType)
			if strings.HasPrefix(ct, MIMEJSON) {
				return json.NewDecoder(r.Body).Decode(v)
			} else if strings.HasPrefix(ct, MIMEForm) {
				return nil
			}
			return ErrUnsupportedMediaType
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

// NotFoundHandler registers a custom NotFound handler.
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

func (e *Echo) add(method, path string, h Handler) {
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
	h, echo := e.Router.Find(r.Method, r.URL.Path, c.params)
	if echo != nil {
		e = echo
	}
	if h == nil {
		h = e.notFoundHandler
	}
	c.reset(w, r, e)

	// Middleware
	var err error
	for i := len(e.middleware) - 1; i >= 0; i-- {
		if h, err = e.middleware[i](h); err != nil {
			e.httpErrorHandler(err, c)
			return
		}
	}

	// Handler
	if err := h(c); err != nil {
		e.httpErrorHandler(err, c)
	}

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
		return func(h HandlerFunc) (HandlerFunc, error) {
			return func(c *Context) error {
				m(c)
				return h(c)
			}, nil
		}
	case func(*Context) error:
		return func(h HandlerFunc) (HandlerFunc, error) {
			var err error
			return func(c *Context) error {
				err = m(c)
				return h(c)
			}, err
		}
	case func(HandlerFunc) (HandlerFunc, error):
		return MiddlewareFunc(m)
	case func(http.Handler) http.Handler:
		return func(h HandlerFunc) (HandlerFunc, error) {
			return func(c *Context) error {
				m(h).ServeHTTP(c.Response, c.Request)
				return h(c)
			}, nil
		}
	case http.Handler, http.HandlerFunc:
		return func(h HandlerFunc) (HandlerFunc, error) {
			return func(c *Context) error {
				m.(http.Handler).ServeHTTP(c.Response, c.Request)
				return h(c)
			}, nil
		}
	case func(http.ResponseWriter, *http.Request):
		return func(h HandlerFunc) (HandlerFunc, error) {
			return func(c *Context) error {
				m(c.Response, c.Request)
				return h(c)
			}, nil
		}
	case func(http.ResponseWriter, *http.Request) error:
		return func(h HandlerFunc) (HandlerFunc, error) {
			var err error
			return func(c *Context) error {
				err = m(c.Response, c.Request)
				return h(c)
			}, err
		}
	default:
		panic("echo: unknown middleware")
	}
}

// wraps Handler
func wrapH(h Handler) HandlerFunc {
	switch h := h.(type) {
	case func(*Context):
		return func(c *Context) error {
			h(c)
			return nil
		}
	case func(*Context) error:
		return HandlerFunc(h)
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
