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
	Server struct {
		config  *engine.Config
		handler engine.Handler
		logger  *log.Logger
		pool    *Pool
	}

	Pool struct {
		request        sync.Pool
		response       sync.Pool
		requestHeader  sync.Pool
		responseHeader sync.Pool
		url            sync.Pool
	}
)

func New(addr string) *Server {
	c := &engine.Config{Address: addr}
	return NewFromConfig(c)
}

func NewFromTLS(addr, certfile, keyfile string) *Server {
	c := &engine.Config{
		Address:     addr,
		TLSCertfile: certfile,
		TLSKeyfile:  keyfile,
	}
	return NewFromConfig(c)
}

func NewFromConfig(c *engine.Config) (s *Server) {
	s = &Server{
		config: c,
		pool: &Pool{
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

func (s *Server) SetHandler(h engine.Handler) {
	s.handler = h
}

func (s *Server) SetLogger(l *log.Logger) {
	s.logger = l
}

func (s *Server) Start() {
	handler := func(c *fasthttp.RequestCtx) {
		// Request
		req := s.pool.request.Get().(*Request)
		reqHdr := s.pool.requestHeader.Get().(*RequestHeader)
		reqURL := s.pool.url.Get().(*URL)
		reqHdr.reset(c.Request.Header)
		reqURL.reset(c.URI())
		req.reset(c, reqHdr, reqURL)

		// Response
		res := s.pool.response.Get().(*Response)
		resHdr := s.pool.responseHeader.Get().(*ResponseHeader)
		resHdr.reset(c.Response.Header)
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
	if certfile != "" && keyfile != "" {
		s.logger.Fatal(fasthttp.ListenAndServeTLS(addr, certfile, keyfile, handler))
	} else {
		s.logger.Fatal(fasthttp.ListenAndServe(addr, handler))
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
