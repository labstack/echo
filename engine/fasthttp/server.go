// +build !appengine

package fasthttp

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/logger"
	"github.com/labstack/gommon/log"
	"github.com/valyala/fasthttp"
)

type (
	Server struct {
		*http.Server
		config  *engine.Config
		handler engine.HandlerFunc
		pool    *Pool
		logger  logger.Logger
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
	return NewConfig(c)
}

func NewTLS(addr, certfile, keyfile string) *Server {
	c := &engine.Config{
		Address:     addr,
		TLSCertfile: certfile,
		TLSKeyfile:  keyfile,
	}
	return NewConfig(c)
}

func NewConfig(c *engine.Config) (s *Server) {
	s = &Server{
		Server: new(http.Server),
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
		handler: func(req engine.Request, res engine.Response) {
			s.logger.Warn("handler not set")
		},
		logger: log.New("echo"),
	}
	return
}

func (s *Server) SetHandler(h engine.HandlerFunc) {
	s.handler = h
}

func (s *Server) SetLogger(l logger.Logger) {
	s.logger = l
}

func (s *Server) Start() {
	fasthttp.ListenAndServe(s.config.Address, func(c *fasthttp.RequestCtx) {
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

		s.handler(req, res)

		s.pool.request.Put(req)
		s.pool.requestHeader.Put(reqHdr)
		s.pool.url.Put(reqURL)
		s.pool.response.Put(res)
		s.pool.responseHeader.Put(resHdr)
	})
	s.logger.Fatal(s.ListenAndServe())
}
