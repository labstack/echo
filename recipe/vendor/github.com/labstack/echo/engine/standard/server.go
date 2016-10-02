package standard

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

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
			panic("echo: handler not set, use `Server#SetHandler()` to set it.")
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
		ln, err := net.Listen("tcp", s.config.Address)
		if err != nil {
			return err
		}

		if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
			// TODO: https://github.com/golang/go/commit/d24f446a90ea94b87591bf16228d7d871fec3d92
			config := &tls.Config{}
			if !s.config.DisableHTTP2 {
				config.NextProtos = append(config.NextProtos, "h2")
			}
			config.Certificates = make([]tls.Certificate, 1)
			config.Certificates[0], err = tls.LoadX509KeyPair(s.config.TLSCertFile, s.config.TLSKeyFile)
			if err != nil {
				return err
			}
			s.config.Listener = tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, config)
		} else {
			s.config.Listener = tcpKeepAliveListener{ln.(*net.TCPListener)}
		}
	}

	return s.Serve(s.config.Listener)
}

// Stop implements `engine.Server#Stop` function.
func (s *Server) Stop() error {
	return s.config.Listener.Close()
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
			})).ServeHTTP(res.adapter, req.Request)
			return
		}
	}
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
