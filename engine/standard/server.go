package standard

import (
	"net/http"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
)

type (
	Server struct {
		*http.Server
		config  *engine.Config
		handler engine.HandlerFunc
		pool    *Pool
		logger  echo.Logger
	}

	Pool struct {
		request  sync.Pool
		response sync.Pool
		header   sync.Pool
		url      sync.Pool
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
		pool: &Pool{
			request: sync.Pool{
				New: func() interface{} {
					return &Request{}
				},
			},
			response: sync.Pool{
				New: func() interface{} {
					return &Response{logger: e.Logger()}
				},
			},
			header: sync.Pool{
				New: func() interface{} {
					return &Header{}
				},
			},
			url: sync.Pool{
				New: func() interface{} {
					return &URL{}
				},
			},
		},
		logger: e.Logger(),
	}
}

func (s *Server) Start() {
	s.Addr = s.config.Address
	s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Request
		req := s.pool.request.Get().(*Request)
		reqHdr := s.pool.header.Get().(*Header)
		reqURL := s.pool.url.Get().(*URL)
		reqHdr.reset(r.Header)
		reqURL.reset(r.URL)
		req.reset(r, reqHdr, reqURL)

		// Response
		res := s.pool.response.Get().(*Response)
		resHdr := s.pool.header.Get().(*Header)
		res.reset(w, reqHdr)

		s.handler(req, res)

		s.pool.request.Put(req)
		s.pool.header.Put(reqHdr)
		s.pool.url.Put(reqURL)
		s.pool.response.Put(res)
		s.pool.header.Put(resHdr)
	})
	s.logger.Fatal(s.ListenAndServe())
}
