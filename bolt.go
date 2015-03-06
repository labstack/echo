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
	MIME_JSON = "application/json"

	HeaderAccept             = "Accept"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentLength      = "Content-Length"
	HeaderContentType        = "Content-Type"
)

var MethodMap = map[string]uint8{
	"CONNECT": 1,
	"DELETE":  2,
	"GET":     3,
	"HEAD":    4,
	"OPTIONS": 5,
	"PATCH":   6,
	"POST":    7,
	"PUT":     8,
	"TRACE":   9,
}

func New(opts ...func(*Bolt)) (b *Bolt) {
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

	// Set options
	for _, o := range opts {
		o(b)
	}

	return
}

func MaxParam(n uint8) func(*Bolt) {
	return func(b *Bolt) {
		b.maxParam = n
	}
}

func NotFoundHandler(h HandlerFunc) func(*Bolt) {
	return func(b *Bolt) {
		b.notFoundHandler = h
	}
}

func MethodNotAllowedHandler(h HandlerFunc) func(*Bolt) {
	return func(b *Bolt) {
		b.methodNotAllowedHandler = h
	}
}

func InternalServerErrorHandler(h HandlerFunc) func(*Bolt) {
	return func(b *Bolt) {
		b.internalServerErrorHandler = h
	}
}

// Use adds middleware(s) in the chain.
func (b *Bolt) Use(h ...HandlerFunc) {
	b.handlers = append(b.handlers, h...)
}

func (b *Bolt) Connect(path string, h ...HandlerFunc) {
	b.Handle("CONNECT", path, h)
}

func (b *Bolt) Delete(path string, h ...HandlerFunc) {
	b.Handle("DELETE", path, h)
}

func (b *Bolt) Get(path string, h ...HandlerFunc) {
	b.Handle("GET", path, h)
}

func (b *Bolt) Head(path string, h ...HandlerFunc) {
	b.Handle("HEAD", path, h)
}

func (b *Bolt) Options(path string, h ...HandlerFunc) {
	b.Handle("OPTIONS", path, h)
}

func (b *Bolt) Patch(path string, h ...HandlerFunc) {
	b.Handle("PATCH", path, h)
}

func (b *Bolt) Post(path string, h ...HandlerFunc) {
	b.Handle("POST", path, h)
}

func (b *Bolt) Put(path string, h ...HandlerFunc) {
	b.Handle("PUT", path, h)
}

func (b *Bolt) Trace(path string, h ...HandlerFunc) {
	b.Handle("TRACE", path, h)
}

func (b *Bolt) Handle(method, path string, h []HandlerFunc) {
	h = append(b.handlers, h...)
	l := len(h)
	b.Router.Add(method, path, func(c *Context) {
		c.handlers = h
		c.l = l
		c.Next()
	})
}

func (b *Bolt) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// Find and execute handler
	h, c, s := b.Router.Find(r.Method, r.URL.Path)
	if h != nil {
		c.reset(rw, r)
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
