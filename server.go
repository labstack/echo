package echo

import (
	stdContext "context"
	"crypto/tls"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	banner = "Echo (v%s). High performance, minimalist Go web framework https://echo.labstack.com"
)

// StartConfig is for creating configured http.Server instance to start serve http(s) requests with given Echo instance
type StartConfig struct {
	// Address for the server to listen on (if not using custom listener)
	Address string

	// ListenerNetwork allows setting listener network (see net.Listen for allowed values)
	// Optional: defaults to "tcp"
	ListenerNetwork string

	// CertFilesystem is file system used to load certificates and keys (if certs/keys are given as paths)
	CertFilesystem fs.FS

	// DisableHTTP2 disables supports for HTTP2 in TLS server
	DisableHTTP2 bool

	// HideBanner does not log Echo banner on server startup
	HideBanner bool

	// HidePort does not log port on server startup
	HidePort bool

	// GracefulContext is context that completion signals graceful shutdown start
	GracefulContext stdContext.Context

	// GracefulTimeout is period which server allows listeners to finish serving ongoing requests. If this time is exceeded process is exited
	// Defaults to 10 seconds
	GracefulTimeout time.Duration

	// OnShutdownError allows customization of what happens when (graceful) server Shutdown method returns an error.
	// Defaults to calling e.logger.Error(err)
	OnShutdownError func(err error)

	// TLSConfigFunc allows modifying TLS configuration before listener is created with it.
	TLSConfigFunc func(tlsConfig *tls.Config)

	// ListenerAddrFunc allows getting listener address before server starts serving requests on listener. Useful when
	// address is set as random (`:0`) port.
	ListenerAddrFunc func(addr net.Addr)

	// BeforeServeFunc allows customizing/accessing server before server starts serving requests on listener.
	BeforeServeFunc func(s *http.Server) error
}

// Start starts a HTTP(s) server.
func (sc StartConfig) Start(e *Echo) error {
	logger := e.Logger
	server := http.Server{
		Handler: e,
		// NB: all http.Server errors will be logged through Logger.Write calls. We could create writer that wraps
		// logger and calls Logger.Error internally when http.Server logs error - atm we will use this naive way.
		ErrorLog: log.New(logger, "", 0),
	}

	var tlsConfig *tls.Config = nil
	if sc.TLSConfigFunc != nil {
		tlsConfig = &tls.Config{}
		configureTLS(&sc, tlsConfig)
		sc.TLSConfigFunc(tlsConfig)
	}

	listener, err := createListener(&sc, tlsConfig)
	if err != nil {
		return err
	}
	return serve(&sc, &server, listener, logger)
}

// StartTLS starts a HTTPS server.
// If `certFile` or `keyFile` is `string` the values are treated as file paths.
// If `certFile` or `keyFile` is `[]byte` the values are treated as the certificate or key as-is.
func (sc StartConfig) StartTLS(e *Echo, certFile, keyFile interface{}) error {
	logger := e.Logger
	s := http.Server{
		Handler: e,
		// NB: all http.Server errors will be logged through Logger.Write calls. We could create writer that wraps
		// logger and calls Logger.Error internally when http.Server logs error - atm we will use this naive way.
		ErrorLog: log.New(logger, "", 0),
	}

	certFs := sc.CertFilesystem
	if certFs == nil {
		certFs = os.DirFS(".")
	}
	cert, err := filepathOrContent(certFile, certFs)
	if err != nil {
		return err
	}
	key, err := filepathOrContent(keyFile, certFs)
	if err != nil {
		return err
	}
	cer, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return err
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
	configureTLS(&sc, tlsConfig)
	if sc.TLSConfigFunc != nil {
		sc.TLSConfigFunc(tlsConfig)
	}

	listener, err := createListener(&sc, tlsConfig)
	if err != nil {
		return err
	}
	return serve(&sc, &s, listener, logger)
}

func serve(sc *StartConfig, server *http.Server, listener net.Listener, logger Logger) error {
	if sc.BeforeServeFunc != nil {
		if err := sc.BeforeServeFunc(server); err != nil {
			return err
		}
	}
	startupGreetings(sc, logger, listener)

	if sc.GracefulContext != nil {
		ctx, cancel := stdContext.WithCancel(sc.GracefulContext)
		defer cancel() // make sure this graceful coroutine will end when serve returns by some other means
		go gracefulShutdown(ctx, sc, server, logger)
	}
	return server.Serve(listener)
}

func configureTLS(sc *StartConfig, tlsConfig *tls.Config) {
	if !sc.DisableHTTP2 {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h2")
	}
}

func createListener(sc *StartConfig, tlsConfig *tls.Config) (net.Listener, error) {
	listenerNetwork := sc.ListenerNetwork
	if listenerNetwork == "" {
		listenerNetwork = "tcp"
	}

	var listener net.Listener
	var err error
	if tlsConfig != nil {
		listener, err = tls.Listen(listenerNetwork, sc.Address, tlsConfig)
	} else {
		listener, err = net.Listen(listenerNetwork, sc.Address)
	}
	if err != nil {
		return nil, err
	}

	if sc.ListenerAddrFunc != nil {
		sc.ListenerAddrFunc(listener.Addr())
	}
	return listener, nil
}

func startupGreetings(sc *StartConfig, logger Logger, listener net.Listener) {
	if !sc.HideBanner {
		bannerText := fmt.Sprintf(banner, Version)
		logger.Write([]byte(bannerText))
	}

	if !sc.HidePort {
		logger.Write([]byte(fmt.Sprintf("http(s) server started on %s", listener.Addr())))
	}
}

func filepathOrContent(fileOrContent interface{}, certFilesystem fs.FS) (content []byte, err error) {
	switch v := fileOrContent.(type) {
	case string:
		return fs.ReadFile(certFilesystem, v)
	case []byte:
		return v, nil
	default:
		return nil, ErrInvalidCertOrKeyType
	}
}

func gracefulShutdown(gracefulCtx stdContext.Context, sc *StartConfig, server *http.Server, logger Logger) {
	<-gracefulCtx.Done() // wait until shutdown context is closed.
	// note: is server if closed by other means this method is still run but is good as no-op

	timeout := sc.GracefulTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	shutdownCtx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		// we end up here when listeners are not shut down within given timeout
		if sc.OnShutdownError != nil {
			sc.OnShutdownError(err)
			return
		}
		logger.Error(fmt.Errorf("failed to shut down server within given timeout: %w", err))
	}
}
