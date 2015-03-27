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

	HeaderAccept             = "Accept"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentLength      = "Content-Length"
	HeaderContentType        = "Content-Type"
)

// New creates a echo instance.
func New() (b *Echo) {
	b = &Echo{
		maxParam: 5,
		notFoundHandler: func(c *Context) {
			http.Error(c.Response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			// c.Halt()
		},
		methodNotAllowedHandler: func(c *Context) {
			http.Error(c.Response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			// c.Halt()
		},
		internalServerErrorHandler: func(c *Context) {
			http.Error(c.Response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			// c.Halt()
		},
	}
	b.Router = NewRouter(b)
	b.pool.New = func() interface{} {
		return &Context{
			Response: &response{},
			params:   make(Params, b.maxParam),
			store:    make(store),
			// i:        -1,
			echo: b,
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
func (b *Echo) MaxParam(n uint8) {
	b.maxParam = n
}

// NotFoundHandler sets a custom NotFound handler.
func (b *Echo) NotFoundHandler(h Handler) {
	b.notFoundHandler = wrapH(h)
}

// MethodNotAllowedHandler sets a custom MethodNotAllowed handler.
func (b *Echo) MethodNotAllowedHandler(h Handler) {
	b.methodNotAllowedHandler = wrapH(h)
}

// InternalServerErrorHandler sets a custom InternalServerError handler.
func (b *Echo) InternalServerErrorHandler(h Handler) {
	b.internalServerErrorHandler = wrapH(h)
}

// Use adds handler to the middleware chain.
func (b *Echo) Use(m ...Middleware) {
	for _, h := range m {
		b.middleware = append(b.middleware, wrapM(h))
	}
}

// Connect adds a CONNECT route > handler to the router.
func (b *Echo) Connect(path string, h Handler) {
	b.Router.Add("CONNECT", path, wrapH(h))
}

// Delete adds a DELETE route > handler to the router.
func (b *Echo) Delete(path string, h Handler) {
	b.Router.Add("DELETE", path, wrapH(h))
}

// Get adds a GET route > handler to the router.
func (b *Echo) Get(path string, h Handler) {
	b.Router.Add("GET", path, wrapH(h))
}

// Head adds a HEAD route > handler to the router.
func (b *Echo) Head(path string, h Handler) {
	b.Router.Add("HEAD", path, wrapH(h))
}

// Options adds an OPTIONS route > handler to the router.
func (b *Echo) Options(path string, h Handler) {
	b.Router.Add("OPTIONS", path, wrapH(h))
}

// Patch adds a PATCH route > handler to the router.
func (b *Echo) Patch(path string, h Handler) {
	b.Router.Add("PATCH", path, wrapH(h))
}

// Post adds a POST route > handler to the router.
func (b *Echo) Post(path string, h Handler) {
	b.Router.Add("POST", path, wrapH(h))
}

// Put adds a PUT route > handler to the router.
func (b *Echo) Put(path string, h Handler) {
	b.Router.Add("PUT", path, wrapH(h))
}

// Trace adds a TRACE route > handler to the router.
func (b *Echo) Trace(path string, h Handler) {
	b.Router.Add("TRACE", path, wrapH(h))
}

// Static serves static files.
func (b *Echo) Static(path, root string) {
	fs := http.StripPrefix(path, http.FileServer(http.Dir(root)))
	b.Get(path+"/*", func(c *Context) {
		fs.ServeHTTP(c.Response, c.Request)
	})
}

// ServeFile serves a file.
func (b *Echo) ServeFile(path, file string) {
	b.Get(path, func(c *Context) {
		http.ServeFile(c.Response, c.Request, file)
	})
}

// Index serves index file.
func (b *Echo) Index(file string) {
	b.ServeFile("/", file)
}

func (b *Echo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h, c, s := b.Router.Find(r.Method, r.URL.Path)
	c.reset(rw, r)
	if h != nil {
		// Middleware
		for i := len(b.middleware) - 1; i >= 0; i-- {
			h = b.middleware[i](h)
		}
		// Handler
		h(c)
	} else {
		if s == NotFound {
			b.notFoundHandler(c)
		} else if s == NotAllowed {
			b.methodNotAllowedHandler(c)
		}
	}
	b.pool.Put(c)
}

func (b *Echo) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, b))
}

// wraps Middleware
func wrapM(m Middleware) MiddlewareFunc {
	switch m := m.(type) {
	case func(HandlerFunc) HandlerFunc:
		return MiddlewareFunc(m)
	case http.HandlerFunc, func(http.ResponseWriter, *http.Request), http.Handler:
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) {
				m.(http.Handler).ServeHTTP(c.Response, c.Request)
				h(c)
			}
		}
	case func(http.Handler) http.Handler:
		return func(h HandlerFunc) HandlerFunc {
			return func(c *Context) {
				m(h).ServeHTTP(c.Response, c.Request)
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
	case http.HandlerFunc, func(http.ResponseWriter, *http.Request), http.Handler:
		return func(c *Context) {
			h.(http.Handler).ServeHTTP(c.Response, c.Request)
		}
	default:
		panic("echo: unknown handler")
	}
}
