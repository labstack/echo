// +build !appengine

package fasthttp

import (
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/log"
	glog "github.com/labstack/gommon/log"
	"github.com/valyala/fasthttp"
)

type (
	// Server implements `engine.Server`.
	Server struct {
		*fasthttp.Server
		config  engine.Config
		handler engine.Handler
		logger  log.Logger
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

// New returns `Server` with provided listen address.
func New(addr string) *Server {
	c := engine.Config{Address: addr}
	return WithConfig(c)
}

// WithTLS returns `Server` with provided TLS config.
func WithTLS(addr, certFile, keyFile string) *Server {
	c := engine.Config{
		Address:     addr,
		TLSCertfile: certFile,
		TLSKeyfile:  keyFile,
	}
	return WithConfig(c)
}

// WithConfig returns `Server` with provided config.
func WithConfig(c engine.Config) (s *Server) {
	s = &Server{
		Server: new(fasthttp.Server),
		config: c,
		pool: &pool{
			request: sync.Pool{
				New: func() interface{} {
					return &Request{logger: s.logger}
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
			s.logger.Error("handler not set, use `SetHandler()` to set it.")
		}),
		logger: glog.New("echo"),
	}
	s.Handler = s.ServeHTTP
	return
}

// SetHandler implements `engine.Server#SetHandler` function.
func (s *Server) SetHandler(h engine.Handler) {
	s.handler = h
}

// SetLogger implements `engine.Server#SetLogger` function.
func (s *Server) SetLogger(l log.Logger) {
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

	// Return to pool
	s.pool.request.Put(req)
	s.pool.requestHeader.Put(reqHdr)
	s.pool.url.Put(reqURL)
	s.pool.response.Put(res)
	s.pool.responseHeader.Put(resHdr)
}

// WrapHandler wraps `fasthttp.RequestHandler` into `echo.HandlerFunc`.
func WrapHandler(h fasthttp.RequestHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request().(*Request)
		res := c.Response().(*Response)
		ctx := req.RequestCtx
		h(ctx)
		res.status = ctx.Response.StatusCode()
		res.size = int64(ctx.Response.Header.ContentLength())
		return nil
	}
}

// WrapMiddleware wraps `func(fasthttp.RequestHandler) fasthttp.RequestHandler`
// into `echo.MiddlewareFunc`
func WrapMiddleware(m func(fasthttp.RequestHandler) fasthttp.RequestHandler) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request().(*Request)
			res := c.Response().(*Response)
			ctx := req.RequestCtx
			m(func(ctx *fasthttp.RequestCtx) {
				next(c)
			})(ctx)
			res.status = ctx.Response.StatusCode()
			res.size = int64(ctx.Response.Header.ContentLength())
			return
		}
	}
}
