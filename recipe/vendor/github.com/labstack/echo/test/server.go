package test

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
)

type (
	Server struct {
		*http.Server
		config  *engine.Config
		handler engine.Handler
		pool    *Pool
		logger  *log.Logger
	}

	Pool struct {
		request  sync.Pool
		response sync.Pool
		header   sync.Pool
		url      sync.Pool
	}
)

func New(addr string) *Server {
	c := &engine.Config{Address: addr}
	return NewConfig(c)
}

func NewTLS(addr, certFile, keyFile string) *Server {
	c := &engine.Config{
		Address:     addr,
		TLSCertFile: certFile,
		TLSKeyFile:  keyFile,
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
		handler: engine.HandlerFunc(func(req engine.Request, res engine.Response) {
			panic("echo: handler not set, use `Server#SetHandler()` to set it.")
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

func (s *Server) Start() error {
	if s.config.Listener == nil {
		return s.startDefaultListener()
	}
	return s.startCustomListener()
}

func (s *Server) Stop() error {
	return nil
}

func (s *Server) startDefaultListener() error {
	c := s.config
	if c.TLSCertFile != "" && c.TLSKeyFile != "" {
		return s.ListenAndServeTLS(c.TLSCertFile, c.TLSKeyFile)
	}
	return s.ListenAndServe()
}

func (s *Server) startCustomListener() error {
	return s.Serve(s.config.Listener)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	resHdr.reset(w.Header())
	res.reset(w, resHdr)

	s.handler.ServeHTTP(req, res)

	s.pool.request.Put(req)
	s.pool.header.Put(reqHdr)
	s.pool.url.Put(reqURL)
	s.pool.response.Put(res)
	s.pool.header.Put(resHdr)
}
