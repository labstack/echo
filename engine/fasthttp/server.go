// +build !appengine

package fasthttp

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/valyala/fasthttp"
)

type (
	Server struct {
		*http.Server
		config  *engine.Config
		handler engine.HandlerFunc
		logger  echo.Logger
	}
)

func New(addr string, e *echo.Echo) *Server {
	c := &engine.Config{Address: addr}
	return NewConfig(c, e)
}

func NewTLS(addr, certfile, keyfile string, e *echo.Echo) *Server {
	c := &engine.Config{
		Address:     addr,
		TLSCertfile: certfile,
		TLSKeyfile:  keyfile,
	}
	return NewConfig(c, e)
}

func NewConfig(c *engine.Config, e *echo.Echo) *Server {
	return &Server{
		Server:  new(http.Server),
		config:  c,
		handler: e.ServeHTTP,
		logger:  e.Logger(),
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
