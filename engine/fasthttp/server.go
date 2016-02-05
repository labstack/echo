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
	fasthttp.ListenAndServe(s.config.Address, func(ctx *fasthttp.RequestCtx) {
		req := &Request{
			context: ctx,
			url:     &URL{ctx.URI()},
			header:  &RequestHeader{ctx.Request.Header},
		}
		res := &Response{
			context: ctx,
			header:  &ResponseHeader{ctx.Response.Header},
		}
		s.handler(req, res)
	})
	s.logger.Fatal(s.ListenAndServe())
}
