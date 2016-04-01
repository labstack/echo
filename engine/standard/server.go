package standard

import (
	"net/http"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
)

type (
	// Server implements `engine.Server`.
	Server struct {
		*http.Server
		config  engine.Config
		handler engine.Handler
		logger  *log.Logger
		pool    *pool
	}

	pool struct {
		request         sync.Pool
		response        sync.Pool
		responseAdapter sync.Pool
		header          sync.Pool
		url             sync.Pool
	}
)

// New returns an instance of `standard.Server` with provided listen address.
func New(addr string) *Server {
	c := engine.Config{Address: addr}
	return NewFromConfig(c)
}

// NewFromTLS returns an instance of `standard.Server` from TLS config.
func NewFromTLS(addr, certfile, keyfile string) *Server {
	c := engine.Config{
		Address:     addr,
		TLSCertfile: certfile,
		TLSKeyfile:  keyfile,
	}
	return NewFromConfig(c)
}

// NewFromConfig returns an instance of `standard.Server` from config.
func NewFromConfig(c engine.Config) (s *Server) {
	s = &Server{
		Server: new(http.Server),
		config: c,
		pool: &pool{
			request: sync.Pool{
				New: func() interface{} {
					return &Request{logger: s.logger}
				},
			},
			response: sync.Pool{
				New: func() interface{} {
					return &Response{logger: s.logger}
				},
			},
			responseAdapter: sync.Pool{
				New: func() interface{} {
					return &responseAdapter{}
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
		handler: engine.HandlerFunc(func(rq engine.Request, rs engine.Response) {
			s.logger.Error("handler not set, use `SetHandler()` to set it.")
		}),
		logger: log.New("echo"),
	}
	s.Addr = c.Address
	s.Handler = s
	return
}

// SetHandler implements `engine.Server#SetHandler` function.
func (s *Server) SetHandler(h engine.Handler) {
	s.handler = h
}

// SetLogger implements `engine.Server#SetLogger` function.
func (s *Server) SetLogger(l *log.Logger) {
	s.logger = l
}

// Start implements `engine.Server#Start` function.
func (s *Server) Start() error {
	if s.config.Listener == nil {
		return s.startDefaultListener()
	}
	return s.startCustomListener()
}

func (s *Server) startDefaultListener() error {
	c := s.config
	if c.TLSCertfile != "" && c.TLSKeyfile != "" {
		return s.ListenAndServeTLS(c.TLSCertfile, c.TLSKeyfile)
	}
	return s.ListenAndServe()
}

func (s *Server) startCustomListener() error {
	return s.Serve(s.config.Listener)
}

// ServeHTTP implements `http.Handler` interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Request
	rq := s.pool.request.Get().(*Request)
	rqHdr := s.pool.header.Get().(*Header)
	rqURL := s.pool.url.Get().(*URL)
	rqHdr.reset(r.Header)
	rqURL.reset(r.URL)
	rq.reset(r, rqHdr, rqURL)

	// Response
	rs := s.pool.response.Get().(*Response)
	rsAdpt := s.pool.responseAdapter.Get().(*responseAdapter)
	rsAdpt.reset(w, rs)
	rsHdr := s.pool.header.Get().(*Header)
	rsHdr.reset(w.Header())
	rs.reset(w, rsAdpt, rsHdr)

	s.handler.ServeHTTP(rq, rs)

	// Return to pool
	s.pool.request.Put(rq)
	s.pool.header.Put(rqHdr)
	s.pool.url.Put(rqURL)
	s.pool.response.Put(rs)
	s.pool.header.Put(rsHdr)
}

// WrapHandler wraps `http.Handler` into `echo.HandlerFunc`.
func WrapHandler(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		rq := c.Request().(*Request)
		rs := c.Response().(*Response)
		h.ServeHTTP(rs.ResponseWriter, rq.Request)
		return nil
	}
}

// WrapMiddleware wraps `func(http.Handler) http.Handler` into `echo.MiddlewareFunc`
func WrapMiddleware(m func(http.Handler) http.Handler) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) (err error) {
			rq := c.Request().(*Request)
			rs := c.Response().(*Response)
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err = next.Handle(c)
			})).ServeHTTP(rs.ResponseWriter, rq.Request)
			return
		})
	}
}
