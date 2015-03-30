package echo

import (
	"log"
	"net/http"
	"sync"
)

type (
	Echo struct {
		Router                     *router
		middleware                 []MiddlewareFunc
		maxParam                   byte
		notFoundHandler            HandlerFunc
		methodNotAllowedHandler    HandlerFunc
		internalServerErrorHandler HandlerFunc
		pool                       sync.Pool
	}
	Handler        interface{}
	HandlerFunc    func(*Context)
	Middleware     interface{}
	MiddlewareFunc func(HandlerFunc) HandlerFunc
)

const (
	MIMEJSON = "application/json"
	MIMEText = "text/plain"

	HeaderAccept             = "Accept"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentLength      = "Content-Length"
	HeaderContentType        = "Content-Type"
)

// New creates a echo instance.
func New() (e *Echo) {
	e = &Echo{
		maxParam: 5,
		notFoundHandler: func(c *Context) {
			http.Error(c.Response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		},
		methodNotAllowedHandler: func(c *Context) {
			http.Error(c.Response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		},
		internalServerErrorHandler: func(c *Context) {
			http.Error(c.Response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		},
	}
	e.Router = NewRouter(e)
	e.pool.New = func() interface{} {
		return &Context{
			Response: &response{},
			params:   make(Params, e.maxParam),
			store:    make(store),
			echo:     e,
		}
	}
	return
}

// NOP
func (h HandlerFunc) ServeHTTP(r http.ResponseWriter, w *http.Request) {
}

// func (b *Echo) Sub(prefix string, m ...MiddlewareFunc) *Echo {
// 	return &Echo{
// prefix:   b.prefix + prefix,
// middleware: append(b.handlers, handlers...),
// 	}
// }

// MaxParam sets the maximum allowed path parameters. Default is 5, good enough
// for many users.
func (e *Echo) MaxParam(n uint8) {
	e.maxParam = n
}

// NotFoundHandler sets a custom NotFound handler.
func (e *Echo) NotFoundHandler(h Handler) {
	e.notFoundHandler = wrapH(h)
}

// MethodNotAllowedHandler sets a custom MethodNotAllowed handler.
func (e *Echo) MethodNotAllowedHandler(h Handler) {
	e.methodNotAllowedHandler = wrapH(h)
}

// InternalServerErrorHandler sets a custom InternalServerError handler.
func (e *Echo) InternalServerErrorHandler(h Handler) {
	e.internalServerErrorHandler = wrapH(h)
}

// Use adds handler to the middleware chain.
func (e *Echo) Use(m ...Middleware) {
	for _, h := range m {
		e.middleware = append(e.middleware, wrapM(h))
	}
}

// Connect adds a CONNECT route > handler to the router.
func (e *Echo) Connect(path string, h Handler) {
	e.Router.Add("CONNECT", path, wrapH(h))
}

// Delete adds a DELETE route > handler to the router.
func (e *Echo) Delete(path string, h Handler) {
	e.Router.Add("DELETE", path, wrapH(h))
}

// Get adds a GET route > handler to the router.
func (e *Echo) Get(path string, h Handler) {
	e.Router.Add("GET", path, wrapH(h))
}

// Head adds a HEAD route > handler to the router.
func (e *Echo) Head(path string, h Handler) {
	e.Router.Add("HEAD", path, wrapH(h))
}

// Options adds an OPTIONS route > handler to the router.
func (e *Echo) Options(path string, h Handler) {
	e.Router.Add("OPTIONS", path, wrapH(h))
}

// Patch adds a PATCH route > handler to the router.
func (e *Echo) Patch(path string, h Handler) {
	e.Router.Add("PATCH", path, wrapH(h))
}

// Post adds a POST route > handler to the router.
func (e *Echo) Post(path string, h Handler) {
	e.Router.Add("POST", path, wrapH(h))
}

// Put adds a PUT route > handler to the router.
func (e *Echo) Put(path string, h Handler) {
	e.Router.Add("PUT", path, wrapH(h))
}

// Trace adds a TRACE route > handler to the router.
func (e *Echo) Trace(path string, h Handler) {
	e.Router.Add("TRACE", path, wrapH(h))
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

func (e *Echo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h, c, s := e.Router.Find(r.Method, r.URL.Path)
	c.reset(rw, r)
	if h != nil {
		// Middleware
		for i := len(e.middleware) - 1; i >= 0; i-- {
			h = e.middleware[i](h)
		}
		// Handler
		h(c)
	} else {
		if s == NotFound {
			e.notFoundHandler(c)
		} else if s == NotAllowed {
			e.methodNotAllowedHandler(c)
		}
	}
	e.pool.Put(c)
}

func (e *Echo) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, e))
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
