package echo

import (
	"bytes"
	stdContext "context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func startOnRandomPort(ctx stdContext.Context, e *Echo) (string, error) {
	addrChan := make(chan string)
	errCh := make(chan error)

	go func() {
		errCh <- (&StartConfig{
			Address:         ":0",
			GracefulContext: ctx,
			GracefulTimeout: 100 * time.Millisecond,
			ListenerAddrFunc: func(addr net.Addr) {
				addrChan <- addr.String()
			},
		}).Start(e)
	}()

	return waitForServerStart(addrChan, errCh)
}

func waitForServerStart(addrChan <-chan string, errCh <-chan error) (string, error) {
	waitCtx, cancel := stdContext.WithTimeout(stdContext.Background(), 200*time.Millisecond)
	defer cancel()

	// wait for addr to arrive
	for {
		select {
		case <-waitCtx.Done():
			return "", waitCtx.Err()
		case addr := <-addrChan:
			return addr, nil
		case err := <-errCh:
			if err == http.ErrServerClosed { // was closed normally before listener callback was called. should not be possible
				return "", nil
			}
			// failed to start and we did not manage to get even listener part.
			return "", err
		}
	}
}

func doGet(url string) (int, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(body), nil
}

func TestStartConfig_Start(t *testing.T) {
	e := New()
	e.GET("/ok", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})

	addrChan := make(chan string)
	errCh := make(chan error)

	ctx, shutdown := stdContext.WithTimeout(stdContext.Background(), 200*time.Millisecond)
	defer shutdown()
	go func() {
		errCh <- (&StartConfig{
			Address:         ":0",
			GracefulContext: ctx,
			ListenerAddrFunc: func(addr net.Addr) {
				addrChan <- addr.String()
			},
		}).Start(e)
	}()

	addr, err := waitForServerStart(addrChan, errCh)
	assert.NoError(t, err)

	// check if server is actually up
	code, body, err := doGet(fmt.Sprintf("http://%v/ok", addr))
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)

	shutdown()

	<-errCh // we will be blocking here until server returns from http.Serve

	// check if server was stopped
	code, body, err = doGet(fmt.Sprintf("http://%v/ok", addr))
	assert.Equal(t, 0, code)
	assert.Equal(t, "", body)

	if err == nil {
		t.Errorf("missing error")
		return
	}
	expectContains := "connect: connection refused"
	if runtime.GOOS == "windows" {
		expectContains = "No connection could be made"
	}
	assert.True(t, strings.Contains(err.Error(), expectContains))
}

func TestStartConfig_GracefulShutdown(t *testing.T) {
	var testCases = []struct {
		name                   string
		whenHandlerTakesLonger bool
		expectBody             string
		expectGracefulError    string
	}{
		{
			name:                   "ok, all handlers returns before graceful shutdown deadline",
			whenHandlerTakesLonger: false,
			expectBody:             "OK",
			expectGracefulError:    "",
		},
		{
			name:                   "nok, handlers do not returns before graceful shutdown deadline",
			whenHandlerTakesLonger: true,
			expectBody:             "timeout",
			expectGracefulError:    stdContext.DeadlineExceeded.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			e.GET("/ok", func(c Context) error {
				msg := "OK"
				if tc.whenHandlerTakesLonger {
					time.Sleep(150 * time.Millisecond)
					msg = "timeout"
				}
				return c.String(http.StatusOK, msg)
			})

			addrChan := make(chan string)
			errCh := make(chan error)

			ctx, shutdown := stdContext.WithTimeout(stdContext.Background(), 50*time.Millisecond)
			defer shutdown()

			shutdownErrChan := make(chan error, 1)
			go func() {
				errCh <- (&StartConfig{
					Address:         ":0",
					GracefulContext: ctx,
					GracefulTimeout: 50 * time.Millisecond,
					OnShutdownError: func(err error) {
						shutdownErrChan <- err
					},
					ListenerAddrFunc: func(addr net.Addr) {
						addrChan <- addr.String()
					},
				}).Start(e)
			}()

			addr, err := waitForServerStart(addrChan, errCh)
			assert.NoError(t, err)

			code, body, err := doGet(fmt.Sprintf("http://%v/ok", addr))
			if err != nil {
				assert.NoError(t, err)
				return
			}
			assert.Equal(t, http.StatusOK, code)
			assert.Equal(t, tc.expectBody, body)

			var shutdownErr error
			select {
			case shutdownErr = <-shutdownErrChan:
			default:
			}
			if tc.expectGracefulError != "" {
				assert.EqualError(t, shutdownErr, tc.expectGracefulError)
			} else {
				assert.NoError(t, shutdownErr)
			}

			shutdown()

			<-errCh // we will be blocking here until server returns from http.Serve

			// check if server was stopped
			code, body, err = doGet(fmt.Sprintf("http://%v/ok", addr))
			assert.Error(t, err)
			if err != nil {
				expectContains := "connect: connection refused"
				if runtime.GOOS == "windows" {
					expectContains = "No connection could be made"
				}
				assert.True(t, strings.Contains(err.Error(), expectContains))
			}
			assert.Equal(t, 0, code)
			assert.Equal(t, "", body)
		})
	}
}

func TestStartConfig_Start_withTLSConfigFunc(t *testing.T) {
	e := New()

	tlsConfigCalled := false
	s := &StartConfig{
		Address: ":0",
		TLSConfigFunc: func(tlsConfig *tls.Config) {
			tlsConfig.GetCertificate = func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return nil, errors.New("not_implemented")
			}
			tlsConfigCalled = true
		},
		BeforeServeFunc: func(s *http.Server) error {
			return errors.New("stop_now")
		},
	}
	err := s.Start(e)
	assert.EqualError(t, err, "stop_now")
	assert.True(t, tlsConfigCalled)
}

func TestStartConfig_Start_createListenerError(t *testing.T) {
	e := New()

	s := &StartConfig{
		Address: ":0",
		TLSConfigFunc: func(tlsConfig *tls.Config) {
		},
		BeforeServeFunc: func(s *http.Server) error {
			return errors.New("stop_now")
		},
	}
	err := s.Start(e)
	assert.EqualError(t, err, "tls: neither Certificates, GetCertificate, nor GetConfigForClient set in Config")
}

func TestStartConfig_StartTLS(t *testing.T) {
	var testCases = []struct {
		name        string
		addr        string
		certFile    string
		keyFile     string
		expectError string
	}{
		{
			name: "ok",
			addr: ":0",
		},
		{
			name:        "nok, invalid certFile",
			addr:        ":0",
			certFile:    "not existing",
			expectError: "open not existing: no such file or directory",
		},
		{
			name:        "nok, invalid keyFile",
			addr:        ":0",
			keyFile:     "not existing",
			expectError: "open not existing: no such file or directory",
		},
		{
			name:        "nok, failed to create cert out of certFile and keyFile",
			addr:        ":0",
			keyFile:     "_fixture/certs/cert.pem", // we are passing cert instead of key
			expectError: "tls: found a certificate rather than a key in the PEM for the private key",
		},
		{
			name:        "nok, invalid tls address",
			addr:        "nope",
			expectError: "listen tcp: address nope: missing port in address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			addrChan := make(chan string)
			errCh := make(chan error)

			ctx, shutdown := stdContext.WithTimeout(stdContext.Background(), 200*time.Millisecond)
			defer shutdown()
			go func() {
				certFile := "_fixture/certs/cert.pem"
				if tc.certFile != "" {
					certFile = tc.certFile
				}
				keyFile := "_fixture/certs/key.pem"
				if tc.keyFile != "" {
					keyFile = tc.keyFile
				}

				s := &StartConfig{
					Address:         tc.addr,
					GracefulContext: ctx,
					GracefulTimeout: 100 * time.Millisecond,
					ListenerAddrFunc: func(addr net.Addr) {
						addrChan <- addr.String()
					},
				}
				errCh <- s.StartTLS(e, certFile, keyFile)
			}()

			_, err := waitForServerStart(addrChan, errCh)

			if tc.expectError != "" {
				if _, ok := err.(*os.PathError); ok {
					assert.Error(t, err) // error messages for unix and windows are different. so name only error type here
				} else {
					assert.EqualError(t, err, tc.expectError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStartConfig_StartTLS_withTLSConfigFunc(t *testing.T) {
	e := New()

	tlsConfigCalled := false
	s := &StartConfig{
		Address: ":0",
		TLSConfigFunc: func(tlsConfig *tls.Config) {
			assert.Len(t, tlsConfig.Certificates, 1)
			tlsConfigCalled = true
		},
		BeforeServeFunc: func(s *http.Server) error {
			return errors.New("stop_now")
		},
	}
	err := s.StartTLS(e, "_fixture/certs/cert.pem", "_fixture/certs/key.pem")

	assert.EqualError(t, err, "stop_now")
	assert.True(t, tlsConfigCalled)
}

func TestStartConfig_StartTLSAndStart(t *testing.T) {
	// We name if Echo and listeners work correctly when Echo is simultaneously attached to HTTP and HTTPS server
	e := New()
	e.GET("/", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})

	tlsCtx, tlsShutdown := stdContext.WithTimeout(stdContext.Background(), 100*time.Millisecond)
	defer tlsShutdown()
	addrTLSChan := make(chan string)
	errTLSChan := make(chan error)
	go func() {
		s := &StartConfig{
			Address:         ":0",
			GracefulContext: tlsCtx,
			GracefulTimeout: 100 * time.Millisecond,
			ListenerAddrFunc: func(addr net.Addr) {
				addrTLSChan <- addr.String()
			},
		}
		errTLSChan <- s.StartTLS(e, "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
	}()

	tlsAddr, err := waitForServerStart(addrTLSChan, errTLSChan)
	assert.NoError(t, err)

	// check if HTTPS works (note: we are using self signed certs so InsecureSkipVerify=true)
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	res, err := client.Get(fmt.Sprintf("https://%v", tlsAddr))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	ctx, shutdown := stdContext.WithTimeout(stdContext.Background(), 100*time.Millisecond)
	defer shutdown()
	addrChan := make(chan string)
	errChan := make(chan error)
	go func() {
		s := &StartConfig{
			Address:         ":0",
			GracefulContext: ctx,
			GracefulTimeout: 100 * time.Millisecond,
			ListenerAddrFunc: func(addr net.Addr) {
				addrChan <- addr.String()
			},
		}
		errChan <- s.Start(e)
	}()

	addr, err := waitForServerStart(addrChan, errChan)
	assert.NoError(t, err)

	// now we are serving both HTTPS and HTTP listeners. see if HTTP works in addition to HTTPS
	res, err = client.Get(fmt.Sprintf("http://%v", addr))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// see if HTTPS works after HTTP listener is also added
	res, err = client.Get(fmt.Sprintf("https://%v", tlsAddr))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestFilepathOrContent(t *testing.T) {
	cert, err := ioutil.ReadFile("_fixture/certs/cert.pem")
	require.NoError(t, err)
	key, err := ioutil.ReadFile("_fixture/certs/key.pem")
	require.NoError(t, err)

	testCases := []struct {
		name        string
		cert        interface{}
		key         interface{}
		expectedErr error
	}{
		{
			name:        `ValidCertAndKeyFilePath`,
			cert:        "_fixture/certs/cert.pem",
			key:         "_fixture/certs/key.pem",
			expectedErr: nil,
		},
		{
			name:        `ValidCertAndKeyByteString`,
			cert:        cert,
			key:         key,
			expectedErr: nil,
		},
		{
			name:        `InvalidKeyType`,
			cert:        cert,
			key:         1,
			expectedErr: ErrInvalidCertOrKeyType,
		},
		{
			name:        `InvalidCertType`,
			cert:        0,
			key:         key,
			expectedErr: ErrInvalidCertOrKeyType,
		},
		{
			name:        `InvalidCertAndKeyTypes`,
			cert:        0,
			key:         1,
			expectedErr: ErrInvalidCertOrKeyType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			addrChan := make(chan string)
			errCh := make(chan error)

			ctx, shutdown := stdContext.WithTimeout(stdContext.Background(), 200*time.Millisecond)
			defer shutdown()

			go func() {
				s := &StartConfig{
					Address:         ":0",
					CertFilesystem:  os.DirFS("."),
					GracefulContext: ctx,
					GracefulTimeout: 100 * time.Millisecond,
					ListenerAddrFunc: func(addr net.Addr) {
						addrChan <- addr.String()
					},
				}
				errCh <- s.StartTLS(e, tc.cert, tc.key)
			}()

			_, err := waitForServerStart(addrChan, errCh)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func supportsIPv6() bool {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		// Check if any interface has local IPv6 assigned
		if strings.Contains(addr.String(), "::1") {
			return true
		}
	}
	return false
}

func TestStartConfig_WithListenerNetwork(t *testing.T) {
	testCases := []struct {
		name    string
		network string
		address string
	}{
		{
			name:    "tcp ipv4 address",
			network: "tcp",
			address: "127.0.0.1:1323",
		},
		{
			name:    "tcp ipv6 address",
			network: "tcp",
			address: "[::1]:1323",
		},
		{
			name:    "tcp4 ipv4 address",
			network: "tcp4",
			address: "127.0.0.1:1323",
		},
		{
			name:    "tcp6 ipv6 address",
			network: "tcp6",
			address: "[::1]:1323",
		},
	}

	hasIPv6 := supportsIPv6()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !hasIPv6 && strings.Contains(tc.address, "::") {
				t.Skip("Skipping testing IPv6 for " + tc.address + ", not available")
			}

			e := New()
			e.GET("/ok", func(c Context) error {
				return c.String(http.StatusOK, "OK")
			})

			addrChan := make(chan string)
			errCh := make(chan error)

			ctx, shutdown := stdContext.WithTimeout(stdContext.Background(), 200*time.Millisecond)
			defer shutdown()

			go func() {
				s := &StartConfig{
					Address:         tc.address,
					ListenerNetwork: tc.network,
					GracefulContext: ctx,
					GracefulTimeout: 100 * time.Millisecond,
					ListenerAddrFunc: func(addr net.Addr) {
						addrChan <- addr.String()
					},
				}
				errCh <- s.Start(e)
			}()

			_, err := waitForServerStart(addrChan, errCh)
			assert.NoError(t, err)

			code, body, err := doGet(fmt.Sprintf("http://%s/ok", tc.address))
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, code)
			assert.Equal(t, "OK", body)
		})
	}
}

func TestStartConfig_WithHideBanner(t *testing.T) {
	var testCases = []struct {
		name       string
		hideBanner bool
	}{
		{
			name:       "hide banner on startup",
			hideBanner: true,
		},
		{
			name:       "show banner on startup",
			hideBanner: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			buf := new(bytes.Buffer)
			e.Logger = &testLogger{output: buf}

			e.GET("/ok", func(c Context) error {
				return c.String(http.StatusOK, "OK")
			})

			addrChan := make(chan string)
			errCh := make(chan error)

			ctx, shutdown := stdContext.WithTimeout(stdContext.Background(), 200*time.Millisecond)
			defer shutdown()

			go func() {
				_, err := waitForServerStart(addrChan, errCh)
				errCh <- err
				shutdown()
			}()

			s := &StartConfig{
				Address:         ":0",
				HideBanner:      tc.hideBanner,
				GracefulContext: ctx,
				GracefulTimeout: 100 * time.Millisecond,
				ListenerAddrFunc: func(addr net.Addr) {
					addrChan <- addr.String()
				},
			}

			if err := s.Start(e); err != http.ErrServerClosed {
				assert.NoError(t, err)
			}
			assert.NoError(t, <-errCh)

			contains := strings.Contains(buf.String(), "High performance, minimalist Go web framework")
			if tc.hideBanner {
				assert.False(t, contains)
			} else {
				assert.True(t, contains)
			}
		})
	}
}

func TestStartConfig_WithHidePort(t *testing.T) {
	var testCases = []struct {
		name     string
		hidePort bool
	}{
		{
			name:     "hide port on startup",
			hidePort: true,
		},
		{
			name:     "show port on startup",
			hidePort: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			buf := new(bytes.Buffer)
			e.Logger = &testLogger{output: buf}

			e.GET("/ok", func(c Context) error {
				return c.String(http.StatusOK, "OK")
			})

			addrChan := make(chan string)
			errCh := make(chan error, 1)

			ctx, shutdown := stdContext.WithTimeout(stdContext.Background(), 200*time.Millisecond)

			go func() {
				_, err := waitForServerStart(addrChan, errCh)
				errCh <- err
				shutdown()
			}()

			s := &StartConfig{
				Address:         ":0",
				HidePort:        tc.hidePort,
				GracefulContext: ctx,
				GracefulTimeout: 100 * time.Millisecond,
				ListenerAddrFunc: func(addr net.Addr) {
					addrChan <- addr.String()
				},
			}
			if err := s.Start(e); err != http.ErrServerClosed {
				assert.NoError(t, err)
			}
			assert.NoError(t, <-errCh)

			portMsg := fmt.Sprintf("http(s) server started on")
			contains := strings.Contains(buf.String(), portMsg)
			if tc.hidePort {
				assert.False(t, contains)
			} else {
				assert.True(t, contains)
			}
		})
	}
}

func TestStartConfig_WithBeforeServeFunc(t *testing.T) {
	e := New()

	e.GET("/ok", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})

	s := &StartConfig{
		Address: ":0",
		BeforeServeFunc: func(s *http.Server) error {
			return errors.New("is called before serve")
		},
	}
	err := s.Start(e)
	assert.EqualError(t, err, "is called before serve")
}

func TestWithDisableHTTP2(t *testing.T) {
	var testCases = []struct {
		name         string
		disableHTTP2 bool
	}{
		{
			name:         "HTTP2 enabled",
			disableHTTP2: false,
		},
		{
			name:         "HTTP2 disabled",
			disableHTTP2: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			e.GET("/ok", func(c Context) error {
				return c.String(http.StatusOK, "OK")
			})

			addrChan := make(chan string)
			errCh := make(chan error, 1)

			ctx, shutdown := stdContext.WithTimeout(stdContext.Background(), 200*time.Millisecond)
			defer shutdown()

			go func() {
				certFile := "_fixture/certs/cert.pem"
				keyFile := "_fixture/certs/key.pem"

				s := &StartConfig{
					Address:         ":0",
					DisableHTTP2:    tc.disableHTTP2,
					GracefulContext: ctx,
					GracefulTimeout: 100 * time.Millisecond,
					ListenerAddrFunc: func(addr net.Addr) {
						addrChan <- addr.String()
					},
				}
				errCh <- s.StartTLS(e, certFile, keyFile)
			}()

			addr, err := waitForServerStart(addrChan, errCh)
			assert.NoError(t, err)

			url := fmt.Sprintf("https://%v/ok", addr)

			// do ordinary http(s) request
			client := &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}}
			res, err := client.Get(url)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			// do HTTP2 request
			client.Transport = &http2.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			resp, err := client.Get(url)
			if err != nil {
				if tc.disableHTTP2 {
					assert.True(t, strings.Contains(err.Error(), `http2: unexpected ALPN protocol ""; want "h2"`))
					return
				}
				log.Fatalf("Failed get: %s", err)
			}

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("Failed reading response body: %s", err)
			}
			assert.Equal(t, "OK", string(body))

		})
	}
}

type testLogger struct {
	output io.Writer
}

func (l *testLogger) Write(p []byte) (n int, err error) {
	return l.output.Write(p)
}

func (l *testLogger) Printf(format string, args ...interface{}) {
	_, _ = l.output.Write([]byte(fmt.Sprintf(format, args...)))
}

func (l *testLogger) Error(err error) {
	_, _ = l.output.Write([]byte(err.Error()))
}
