package bolt

import (
	"log"
	"net/http"
	"sync"
)

type (
	Bolt struct {
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

// New creates a bolt instance.
func New() (b *Bolt) {
	b = &Bolt{
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
			bolt: b,
		}
	}
	return
}

// NOP
func (h HandlerFunc) ServeHTTP(r http.ResponseWriter, w *http.Request) {
}

func (b *Bolt) Sub(prefix string, m ...MiddlewareFunc) *Bolt {
	return &Bolt{
	// prefix:   b.prefix + prefix,
	// middleware: append(b.handlers, handlers...),
	}
}

// MaxParam sets the max path params allowed. Default is 5, good enough for
// many users.
func (b *Bolt) MaxParam(n uint8) {
	b.maxParam = n
}

// NotFoundHandler sets a custom NotFound handler.
func (b *Bolt) NotFoundHandler(h HandlerFunc) {
	b.notFoundHandler = h
}

// MethodNotAllowedHandler sets a custom MethodNotAllowed handler.
func (b *Bolt) MethodNotAllowedHandler(h HandlerFunc) {
	b.methodNotAllowedHandler = h
}

// InternalServerErrorHandler sets a custom InternalServerError handler.
func (b *Bolt) InternalServerErrorHandler(h HandlerFunc) {
	b.internalServerErrorHandler = h
}

// Use adds handler to the middleware chain.
func (b *Bolt) Use(m ...Middleware) {
	for _, h := range m {
		b.middleware = append(b.middleware, wrapM(h))
	}
}

// Connect adds a CONNECT route.
func (b *Bolt) Connect(path string, h Handler) {
	b.Handle("CONNECT", path, h)
}

// Delete adds a DELETE route.
func (b *Bolt) Delete(path string, h Handler) {
	b.Handle("DELETE", path, h)
}

// Get adds a GET route.
func (b *Bolt) Get(path string, h Handler) {
	b.Handle("GET", path, h)
}

// Head adds a HEAD route.
func (b *Bolt) Head(path string, h Handler) {
	b.Handle("HEAD", path, h)
}

// Options adds an OPTIONS route.
func (b *Bolt) Options(path string, h Handler) {
	b.Handle("OPTIONS", path, h)
}

// Patch adds a PATCH route.
func (b *Bolt) Patch(path string, h Handler) {
	b.Handle("PATCH", path, h)
}

// Post adds a POST route.
func (b *Bolt) Post(path string, h Handler) {
	b.Handle("POST", path, h)
}

// Put adds a PUT route.
func (b *Bolt) Put(path string, h Handler) {
	b.Handle("PUT", path, h)
}

// Trace adds a TRACE route.
func (b *Bolt) Trace(path string, h Handler) {
	b.Handle("TRACE", path, h)
}

// Handle adds method, path  handler to the router.
func (b *Bolt) Handle(method, path string, h Handler) {
	b.Router.Add(method, path, b.wrapH(h))
	// hs := append(b.middleware, wrap(h, false))
	// l := len(hs)
	// b.Router.Add(method, path, func(c *Context) {
	// 	c.handlers = hs
	// 	c.l = l
	// c.Next()
	// })
}

// Static serves static files.
func (b *Bolt) Static(path, root string) {
	fs := http.StripPrefix(path, http.FileServer(http.Dir(root)))
	b.Get(path+"/*", func(c *Context) {
		fs.ServeHTTP(c.Response, c.Request)
	})
}

// ServeFile serves a file.
func (b *Bolt) ServeFile(path, file string) {
	b.Get(path, func(c *Context) {
		http.ServeFile(c.Response, c.Request, file)
	})
}

// Index serves index file.
func (b *Bolt) Index(file string) {
	b.ServeFile("/", file)
}

func (b *Bolt) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
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

func (b *Bolt) Run(addr string) {
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
		panic("bolt: unknown middleware")
	}
}

// wraps Handler
func (b *Bolt) wrapH(h Handler) HandlerFunc {
	switch h := h.(type) {
	case func(*Context):
		return HandlerFunc(h)
	case http.HandlerFunc, func(http.ResponseWriter, *http.Request), http.Handler:
		return func(c *Context) {
			h.(http.Handler).ServeHTTP(c.Response, c.Request)
		}
	default:
		panic("bolt: unknown handler")
	}
}
