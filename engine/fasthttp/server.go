// +build !appengine

package fasthttp

import (
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
	"github.com/valyala/fasthttp"
	"net"
)

type (
	// Server implements `engine.Engine`.
	Server struct {
		config   engine.Config
		handler  engine.Handler
		listener net.Listener
		logger   *log.Logger
		pool     *pool
	}

	pool struct {
		request        sync.Pool
		response       sync.Pool
		requestHeader  sync.Pool
		responseHeader sync.Pool
		url            sync.Pool
	}
)

// New returns an instance of `fasthttp.Server` with provided listen address.
func New(addr string) *Server {
	c := engine.Config{Address: addr}
	return NewFromConfig(c)
}

// NewFromTLS returns an instance of `fasthttp.Server` from TLS config.
func NewFromTLS(addr, certfile, keyfile string) *Server {
	c := engine.Config{
		Address:     addr,
		TLSCertfile: certfile,
		TLSKeyfile:  keyfile,
	}
	return NewFromConfig(c)
}

// NewFromConfig returns an instance of `standard.Server` from config.
func NewFromConfig(c engine.Config) (s *Server) {
	s = &Server{
		config: c,
		pool: &pool{
			request: sync.Pool{
				New: func() interface{} {
					return &Request{}
				},
			},
			response: sync.Pool{
				New: func() interface{} {
					return &Response{logger: s.logger}
				},
			},
			requestHeader: sync.Pool{
				New: func() interface{} {
					return &RequestHeader{}
				},
			},
			responseHeader: sync.Pool{
				New: func() interface{} {
					return &ResponseHeader{}
				},
			},
			url: sync.Pool{
				New: func() interface{} {
					return &URL{}
				},
			},
		},
		handler: engine.HandlerFunc(func(req engine.Request, res engine.Response) {
			s.logger.Fatal("handler not set")
		}),
		logger: log.New("echo"),
	}
	return
}

// SetHandler implements `engine.Engine#SetHandler` method.
func (s *Server) SetHandler(h engine.Handler) {
	s.handler = h
}

// SetHandler implements `engine.Engine#SetListener` method.
func (s *Server) SetListener(ln net.Listener) {
	s.listener = ln
}

// SetLogger implements `engine.Engine#SetLogger` method.
func (s *Server) SetLogger(l *log.Logger) {
	s.logger = l
}

// Start implements `engine.Engine#Start` method.
func (s *Server) Start() {
	handler := func(c *fasthttp.RequestCtx) {
		// Request
		req := s.pool.request.Get().(*Request)
		reqHdr := s.pool.requestHeader.Get().(*RequestHeader)
		reqURL := s.pool.url.Get().(*URL)
		reqHdr.reset(&c.Request.Header)
		reqURL.reset(c.URI())
		req.reset(c, reqHdr, reqURL)

		// Response
		res := s.pool.response.Get().(*Response)
		resHdr := s.pool.responseHeader.Get().(*ResponseHeader)
		resHdr.reset(&c.Response.Header)
		res.reset(c, resHdr)

		s.handler.ServeHTTP(req, res)

		s.pool.request.Put(req)
		s.pool.requestHeader.Put(reqHdr)
		s.pool.url.Put(reqURL)
		s.pool.response.Put(res)
		s.pool.responseHeader.Put(resHdr)
	}

	addr := s.config.Address
	certfile := s.config.TLSCertfile
	keyfile := s.config.TLSKeyfile

	if nil == s.listener {
		s.startDefaultListener(addr, certfile, keyfile, handler)
	} else {
		s.startCustomListener(certfile, keyfile, handler)
	}
}

func (s *Server) startDefaultListener(addr, certfile, keyfile string, handler func(c *fasthttp.RequestCtx)) {
	if certfile != "" && keyfile != "" {
		s.logger.Fatal(fasthttp.ListenAndServeTLS(addr, certfile, keyfile, handler))
	} else {
		s.logger.Fatal(fasthttp.ListenAndServe(addr, handler))
	}
}

func (s *Server) startCustomListener(certfile, keyfile string, handler func(c *fasthttp.RequestCtx)) {
	if certfile != "" && keyfile != "" {
		s.logger.Fatal(fasthttp.ServeTLS(s.listener, certfile, keyfile, handler))
	} else {
		s.logger.Fatal(fasthttp.Serve(s.listener, handler))
	}
}

// WrapHandler wraps `fasthttp.RequestHandler` into `echo.HandlerFunc`.
func WrapHandler(h fasthttp.RequestHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().(*Request).RequestCtx
		h(ctx)
		return nil
	}
}

// WrapMiddleware wraps `fasthttp.RequestHandler` into `echo.MiddlewareFunc`
func WrapMiddleware(h fasthttp.RequestHandler) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			ctx := c.Request().(*Request).RequestCtx
			h(ctx)
			return next.Handle(c)
		})
	}
}
