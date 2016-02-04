package standard

import (
	"log"
	"net/http"
	"sync"

	"github.com/labstack/echo/engine"
)

type (
	Server struct {
		*http.Server
		config  *engine.Config
		handler engine.HandlerFunc
		pool    *Pool
	}

	Pool struct {
		request  sync.Pool
		response sync.Pool
		header   sync.Pool
		url      sync.Pool
	}
)

func NewServer(config *engine.Config, handler engine.HandlerFunc) *Server {
	return &Server{
		Server:  new(http.Server),
		config:  config,
		handler: handler,
		pool: &Pool{
			request: sync.Pool{
				New: func() interface{} {
					return &Request{}
				},
			},
			response: sync.Pool{
				New: func() interface{} {
					return &Response{}
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
	}
}

func (s *Server) Start() {
	s.Addr = s.config.Address
	s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Request
		req := s.pool.request.Get().(*Request)
		reqHdr := s.pool.request.Get().(*Header)
		reqURL := s.pool.request.Get().(*URL)
		reqHdr.reset(r.Header)
		reqURL.reset(r.URL)
		req.reset(r, reqHdr, reqURL)

		// Response
		res := s.pool.request.Get().(*Response)
		resHdr := s.pool.request.Get().(*Header)
		res.reset(w, reqHdr)

		s.handler(req, res)

		s.pool.request.Put(req)
		s.pool.header.Put(reqHdr)
		s.pool.url.Put(reqURL)
		s.pool.response.Put(res)
		s.pool.header.Put(resHdr)
	})
	log.Fatal(s.ListenAndServe())
}
