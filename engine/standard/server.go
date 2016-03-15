package standard

import (
	"net/http"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/gommon/log"
)

type (
	// Server implements `engine.Engine`.
	Server struct {
		*http.Server
		config  engine.Config
		handler engine.Handler
		logger  *log.Logger
		pool    *pool
	}

	pool struct {
		request  sync.Pool
		response sync.Pool
		header   sync.Pool
		url      sync.Pool
	}
)

// New returns an instance of `standard.Server` with specified listen address.
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
			s.logger.Fatal("handler not set")
		}),
		logger: log.New("echo"),
	}
	s.Addr = c.Address
	s.Handler = s
	return
}

// SetHandler implements `engine.Engine#SetHandler` method.
func (s *Server) SetHandler(h engine.Handler) {
	s.handler = h
}

// SetLogger implements `engine.Engine#SetLogger` method.
func (s *Server) SetLogger(l *log.Logger) {
	s.logger = l
}

// Start implements `engine.Engine#Start` method.
func (s *Server) Start() {
	certfile := s.config.TLSCertfile
	keyfile := s.config.TLSKeyfile
	if certfile != "" && keyfile != "" {
		s.logger.Fatal(s.ListenAndServeTLS(certfile, keyfile))
	} else {
		s.logger.Fatal(s.ListenAndServe())
	}
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

// WrapHandler wraps `http.Handler` into `echo.HandlerFunc`.
func WrapHandler(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		w := c.Response().(*Response).ResponseWriter
		r := c.Request().(*Request).Request
		h.ServeHTTP(w, r)
		return nil
	}
}

// WrapMiddleware wraps `func(http.Handler) http.Handler` into `echo.MiddlewareFunc`
func WrapMiddleware(m func(http.Handler) http.Handler) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) (err error) {
			req := c.Request().(*Request)
			res := c.Response().(*Response)
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				res.ResponseWriter = w
				req.Request = r
				err = next.Handle(c)
			})).ServeHTTP(res.ResponseWriter, req.Request)
			return
		})
	}
}
