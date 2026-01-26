// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	stdContext "context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
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
	// Address specifies the address where listener will start listening on to serve HTTP(s) requests
	Address string

	// HideBanner instructs Start* method not to print banner when starting the Server.
	HideBanner bool
	// HidePort instructs Start* method not to print port when starting the Server.
	HidePort bool

	// CertFilesystem is filesystem is used to read `certFile` and `keyFile` when StartTLS method is called.
	CertFilesystem fs.FS
	TLSConfig      *tls.Config

	// ListenerNetwork is used configure on which Network listener will use.
	ListenerNetwork string
	// ListenerAddrFunc will be called after listener is created and started to listen for connections. This is useful in
	// testing situations when server is started on random port `addres = ":0"` in that case you can get actual port where
	// listener is listening on.
	ListenerAddrFunc func(addr net.Addr)

	// GracefulTimeout is timeout value (defaults to 10sec) graceful shutdown will wait for server to handle ongoing requests
	// before shutting down the server.
	GracefulTimeout time.Duration
	// OnShutdownError is called when graceful shutdown results an error. for example when listeners are not shut down within
	// given timeout
	OnShutdownError func(err error)

	// BeforeServeFunc is callback that is called just before server starts to serve HTTP request.
	// Use this callback when you want to configure http.Server different timeouts/limits/etc
	BeforeServeFunc func(s *http.Server) error
}

// Start starts given Handler with HTTP(s) server.
func (sc StartConfig) Start(ctx stdContext.Context, h http.Handler) error {
	return sc.start(ctx, h)
}

// StartTLS starts given Handler with HTTPS server.
// If `certFile` or `keyFile` is `string` the values are treated as file paths.
// If `certFile` or `keyFile` is `[]byte` the values are treated as the certificate or key as-is.
func (sc StartConfig) StartTLS(ctx stdContext.Context, h http.Handler, certFile, keyFile any) error {
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
	if sc.TLSConfig == nil {
		sc.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"h2"},
			//NextProtos: []string{"http/1.1"}, // Disallow "h2", allow http
		}
	}
	sc.TLSConfig.Certificates = []tls.Certificate{cer}
	return sc.start(ctx, h)
}

// start starts handler with HTTP(s) server.
func (sc StartConfig) start(ctx stdContext.Context, h http.Handler) error {
	var logger *slog.Logger
	if e, ok := h.(*Echo); ok {
		logger = e.Logger
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	server := http.Server{
		Handler:  h,
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		// defaults for GoSec rule G112 // https://github.com/securego/gosec
		// G112 (CWE-400): Potential Slowloris Attack because ReadHeaderTimeout is not configured in the http.Server
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	listenerNetwork := sc.ListenerNetwork
	if listenerNetwork == "" {
		listenerNetwork = "tcp"
	}
	var listener net.Listener
	var err error
	if sc.TLSConfig != nil {
		listener, err = tls.Listen(listenerNetwork, sc.Address, sc.TLSConfig)
	} else {
		listener, err = net.Listen(listenerNetwork, sc.Address)
	}
	if err != nil {
		return err
	}
	if sc.ListenerAddrFunc != nil {
		sc.ListenerAddrFunc(listener.Addr())
	}

	if sc.BeforeServeFunc != nil {
		if err := sc.BeforeServeFunc(&server); err != nil {
			_ = listener.Close()
			return err
		}
	}
	if !sc.HideBanner {
		bannerText := fmt.Sprintf(banner, Version)
		logger.Info(bannerText, "version", Version)
	}
	if !sc.HidePort {
		logger.Info("http(s) server started", "address", listener.Addr().String())
	}

	if sc.GracefulTimeout >= 0 {
		gCtx, cancel := stdContext.WithCancel(ctx) // end goroutine when Serve returns early
		defer cancel()
		go gracefulShutdown(gCtx, &sc, &server, logger)
	}

	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func filepathOrContent(fileOrContent any, certFilesystem fs.FS) (content []byte, err error) {
	switch v := fileOrContent.(type) {
	case string:
		return fs.ReadFile(certFilesystem, v)
	case []byte:
		return v, nil
	default:
		return nil, ErrInvalidCertOrKeyType
	}
}

func gracefulShutdown(shutdownCtx stdContext.Context, sc *StartConfig, server *http.Server, logger *slog.Logger) {
	<-shutdownCtx.Done() // wait until shutdown context is closed.
	// note: is server if closed by other means this method is still run but is good as no-op

	timeout := sc.GracefulTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	waitShutdownCtx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(waitShutdownCtx); err != nil {
		// we end up here when listeners are not shut down within given timeout
		if sc.OnShutdownError != nil {
			sc.OnShutdownError(err)
			return
		}
		logger.Error("failed to shut down server within given timeout", "error", err)
	}
}
