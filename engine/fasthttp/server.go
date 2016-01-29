package fasthttp

import (
	"log"
	"net/http"
)
import (
	"github.com/labstack/echo/engine"
	"github.com/valyala/fasthttp"
)

type (
	Server struct {
		*http.Server
		config  *engine.Config
		handler engine.HandlerFunc
	}
)

func NewServer(config *engine.Config, handler engine.HandlerFunc) *Server {
	return &Server{
		Server:  new(http.Server),
		config:  config,
		handler: handler,
	}
}

func (s *Server) Start() {
	fasthttp.ListenAndServe(s.config.Address, func(ctx *fasthttp.RequestCtx) {
		println("FastHTTP")
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
	log.Fatal(s.ListenAndServe())
}
