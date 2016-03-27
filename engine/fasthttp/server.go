// +build !appengine

package fasthttp

import (
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
	"github.com/valyala/fasthttp"
)

type (
	// Server implements `engine.Server`.
	Server struct {
		*fasthttp.Server
		config  engine.Config
		handler engine.Handler
		logger  *log.Logger
		pool    *pool
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
		Server: new(fasthttp.Server),
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
		handler: engine.HandlerFunc(func(rq engine.Request, rs engine.Response) {
			s.logger.Error("handler not set, use `SetHandler()` to set it.")
		}),
		logger: log.New("echo"),
	}
	s.Handler = s.ServeHTTP
	return
}

// SetHandler implements `engine.Server#SetHandler` function.
func (s *Server) SetHandler(h engine.Handler) {
	s.handler = h
}

// SetLogger implements `engine.Server#SetLogger` function.
func (s *Server) SetLogger(l *log.Logger) {
	s.logger = l
}

// Start implements `engine.Server#Start` function.
func (s *Server) Start() error {
	if s.config.Listener == nil {
		return s.startDefaultListener()
	}
	return s.startCustomListener()

}

func (s *Server) startDefaultListener() error {
	c := s.config
	if c.TLSCertfile != "" && c.TLSKeyfile != "" {
		return s.ListenAndServeTLS(c.Address, c.TLSCertfile, c.TLSKeyfile)
	}
	return s.ListenAndServe(c.Address)
}

func (s *Server) startCustomListener() error {
	c := s.config
	if c.TLSCertfile != "" && c.TLSKeyfile != "" {
		return s.ServeTLS(c.Listener, c.TLSCertfile, c.TLSKeyfile)
	}
	return s.Serve(c.Listener)
}

func (s *Server) ServeHTTP(c *fasthttp.RequestCtx) {
	// Request
	rq := s.pool.request.Get().(*Request)
	reqHdr := s.pool.requestHeader.Get().(*RequestHeader)
	reqURL := s.pool.url.Get().(*URL)
	reqHdr.reset(&c.Request.Header)
	reqURL.reset(c.URI())
	rq.reset(c, reqHdr, reqURL)

	// Response
	rs := s.pool.response.Get().(*Response)
	resHdr := s.pool.responseHeader.Get().(*ResponseHeader)
	resHdr.reset(&c.Response.Header)
	rs.reset(c, resHdr)

	s.handler.ServeHTTP(rq, rs)

	s.pool.request.Put(rq)
	s.pool.requestHeader.Put(reqHdr)
	s.pool.url.Put(reqURL)
	s.pool.response.Put(rs)
	s.pool.responseHeader.Put(resHdr)
}

// WrapHandler wraps `fasthttp.RequestHandler` into `echo.HandlerFunc`.
func WrapHandler(h fasthttp.RequestHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		rq := c.Request().(*Request)
		rs := c.Response().(*Response)
		ctx := rq.RequestCtx
		h(ctx)
		rs.status = ctx.Response.StatusCode()
		rs.size = int64(ctx.Response.Header.ContentLength())
		return nil
	}
}

// WrapMiddleware wraps `fasthttp.RequestHandler` into `echo.MiddlewareFunc`
func WrapMiddleware(h fasthttp.RequestHandler) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			rq := c.Request().(*Request)
			rs := c.Response().(*Response)
			ctx := rq.RequestCtx
			h(ctx)
			rs.status = ctx.Response.StatusCode()
			rs.size = int64(ctx.Response.Header.ContentLength())
			return next.Handle(c)
		})
	}
}
