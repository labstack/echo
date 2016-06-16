package standard

import (
	"net/http"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/log"
	glog "github.com/labstack/gommon/log"
)

type (
	// Server implements `engine.Server`.
	Server struct {
		*http.Server
		config  engine.Config
		handler engine.Handler
		logger  log.Logger
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

// New returns `Server` instance with provided listen address.
func New(addr string) *Server {
	c := engine.Config{Address: addr}
	return WithConfig(c)
}

// WithTLS returns `Server` instance with provided TLS config.
func WithTLS(addr, certFile, keyFile string) *Server {
	c := engine.Config{
		Address:     addr,
		TLSCertFile: certFile,
		TLSKeyFile:  keyFile,
	}
	return WithConfig(c)
}

// WithConfig returns `Server` instance with provided config.
func WithConfig(c engine.Config) (s *Server) {
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
		handler: engine.HandlerFunc(func(req engine.Request, res engine.Response) {
			s.logger.Error("handler not set, use `SetHandler()` to set it.")
		}),
		logger: glog.New("echo"),
	}
	s.ReadTimeout = c.ReadTimeout
	s.WriteTimeout = c.WriteTimeout
	s.Addr = c.Address
	s.Handler = s
	return
}

// SetHandler implements `engine.Server#SetHandler` function.
func (s *Server) SetHandler(h engine.Handler) {
	s.handler = h
}

// SetLogger implements `engine.Server#SetLogger` function.
func (s *Server) SetLogger(l log.Logger) {
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
	if c.TLSCertFile != "" && c.TLSKeyFile != "" {
		return s.ListenAndServeTLS(c.TLSCertFile, c.TLSKeyFile)
	}
	return s.ListenAndServe()
}

func (s *Server) startCustomListener() error {
	return s.Serve(s.config.Listener)
}

// ServeHTTP implements `http.Handler` interface.
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
	resAdpt := s.pool.responseAdapter.Get().(*responseAdapter)
	resAdpt.reset(res)
	resHdr := s.pool.header.Get().(*Header)
	resHdr.reset(w.Header())
	res.reset(w, resAdpt, resHdr)

	s.handler.ServeHTTP(req, res)

	// Return to pool
	s.pool.request.Put(req)
	s.pool.header.Put(reqHdr)
	s.pool.url.Put(reqURL)
	s.pool.response.Put(res)
	s.pool.header.Put(resHdr)
}

// WrapHandler wraps `http.Handler` into `echo.HandlerFunc`.
func WrapHandler(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request().(*Request)
		res := c.Response().(*Response)
		h.ServeHTTP(res.adapter, req.Request)
		return nil
	}
}

// WrapMiddleware wraps `func(http.Handler) http.Handler` into `echo.MiddlewareFunc`
func WrapMiddleware(m func(http.Handler) http.Handler) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request().(*Request)
			res := c.Response().(*Response)
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err = next(c)
			})).ServeHTTP(res.ResponseWriter, req.Request)
			return
		}
	}
}
