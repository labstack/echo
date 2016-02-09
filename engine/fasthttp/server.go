// +build !appengine

package fasthttp

import (
	"net/http"

	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/logger"
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

func NewServer(c *engine.Config, h engine.HandlerFunc, l logger.Logger) *Server {
	return &Server{
		Server:  new(http.Server),
		config:  c,
		handler: h,
		logger:  l,
	}
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
