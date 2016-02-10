// +build !appengine

package fasthttp

import (
	"net/http"

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
		logger  logger.Logger
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
		handler: func(req engine.Request, res engine.Response) {
			s.logger.Info("handler not set")
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
		req := &Request{
			context: c,
			url:     &URL{c.URI()},
			header:  &RequestHeader{c.Request.Header},
		}
		res := &Response{
			context: c,
			header:  &ResponseHeader{c.Response.Header},
		}
		s.handler(req, res)
	})
	s.logger.Fatal(s.ListenAndServe())
}
