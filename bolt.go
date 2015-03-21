package bolt

import (
	"log"
	"net/http"
	"sync"
)

type (
	Bolt struct {
		Router                     *router
		handlers                   []HandlerFunc
		maxParam                   byte
		notFoundHandler            HandlerFunc
		methodNotAllowedHandler    HandlerFunc
		internalServerErrorHandler HandlerFunc
		pool                       sync.Pool
	}
	HandlerFunc func(*Context)
)

const (
	MIMEJSON = "application/json"

	HeaderAccept             = "Accept"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentLength      = "Content-Length"
	HeaderContentType        = "Content-Type"
)

// Methods is a map for looking up HTTP method index.
var Methods = map[string]uint8{
	"CONNECT": 0,
	"DELETE":  1,
	"GET":     2,
	"HEAD":    3,
	"OPTIONS": 4,
	"PATCH":   5,
	"POST":    6,
	"PUT":     7,
	"TRACE":   8,
}

// New creates a bolt instance.
func New() (b *Bolt) {
	b = &Bolt{
		maxParam: 5,
		notFoundHandler: func(c *Context) {
			http.Error(c.Response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			c.Halt()
		},
		methodNotAllowedHandler: func(c *Context) {
			http.Error(c.Response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			c.Halt()
		},
		internalServerErrorHandler: func(c *Context) {
			http.Error(c.Response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			c.Halt()
		},
	}
	b.Router = NewRouter(b)
	b.pool.New = func() interface{} {
		return &Context{
			Response: &response{},
			params:   make(Params, b.maxParam),
			store:    make(store),
			i:        -1,
			bolt:     b,
		}
	}
	return
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

// Chain adds middleware to the chain.
func (b *Bolt) Chain(h ...HandlerFunc) {
	b.handlers = append(b.handlers, h...)
}

// Wrap wraps any http.Handler into bolt.HandlerFunc. It facilitates to use
// third party handler / middleware with bolt.
func (b *Bolt) Wrap(h http.Handler) HandlerFunc {
	return func(c *Context) {
		h.ServeHTTP(c.Response, c.Request)
		c.Next()
	}
}

// Connect adds a CONNECT route.
func (b *Bolt) Connect(path string, h ...HandlerFunc) {
	b.Handle("CONNECT", path, h)
}

// Delete adds a DELETE route.
func (b *Bolt) Delete(path string, h ...HandlerFunc) {
	b.Handle("DELETE", path, h)
}

// Get adds a GET route.
func (b *Bolt) Get(path string, h ...HandlerFunc) {
	b.Handle("GET", path, h)
}

// Head adds a HEAD route.
func (b *Bolt) Head(path string, h ...HandlerFunc) {
	b.Handle("HEAD", path, h)
}

// Options adds an OPTIONS route.
func (b *Bolt) Options(path string, h ...HandlerFunc) {
	b.Handle("OPTIONS", path, h)
}

// Patch adds a PATCH route.
func (b *Bolt) Patch(path string, h ...HandlerFunc) {
	b.Handle("PATCH", path, h)
}

// Post adds a POST route.
func (b *Bolt) Post(path string, h ...HandlerFunc) {
	b.Handle("POST", path, h)
}

// Put adds a PUT route.
func (b *Bolt) Put(path string, h ...HandlerFunc) {
	b.Handle("PUT", path, h)
}

// Trace adds a TRACE route.
func (b *Bolt) Trace(path string, h ...HandlerFunc) {
	b.Handle("TRACE", path, h)
}

// Handle adds method, path  handler to the router.
func (b *Bolt) Handle(method, path string, h []HandlerFunc) {
	h = append(b.handlers, h...)
	l := len(h)
	b.Router.Add(method, path, func(c *Context) {
		c.handlers = h
		c.l = l
		c.Next()
	})
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
	// Find and execute handler
	h, c, s := b.Router.Find(r.Method, r.URL.Path)
	c.reset(rw, r)
	if h != nil {
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

func (b *Bolt) Stop(addr string) {
	panic("implement it")
}
