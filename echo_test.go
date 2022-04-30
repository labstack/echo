package echo

import (
	"bytes"
	stdContext "context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

type (
	user struct {
		ID   int    `json:"id" xml:"id" form:"id" query:"id" param:"id" header:"id"`
		Name string `json:"name" xml:"name" form:"name" query:"name" param:"name" header:"name"`
	}
)

const (
	userJSON                    = `{"id":1,"name":"Jon Snow"}`
	usersJSON                   = `[{"id":1,"name":"Jon Snow"}]`
	userXML                     = `<user><id>1</id><name>Jon Snow</name></user>`
	userForm                    = `id=1&name=Jon Snow`
	invalidContent              = "invalid content"
	userJSONInvalidType         = `{"id":"1","name":"Jon Snow"}`
	userXMLConvertNumberError   = `<user><id>Number one</id><name>Jon Snow</name></user>`
	userXMLUnsupportedTypeError = `<user><>Number one</><name>Jon Snow</name></user>`
)

const userJSONPretty = `{
  "id": 1,
  "name": "Jon Snow"
}`

const userXMLPretty = `<user>
  <id>1</id>
  <name>Jon Snow</name>
</user>`

var dummyQuery = url.Values{"dummy": []string{"useless"}}

func TestEcho(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// DefaultHTTPErrorHandler
	e.DefaultHTTPErrorHandler(errors.New("error"), c)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEchoMiddleware(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.String(http.StatusOK, "OK")
		}
	})

	c, b := request(http.MethodGet, "/", e)
	assert.Equal(t, "123", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoMiddlewareError(t *testing.T) {
	e := New()
	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return errors.New("error")
		}
	})
	c, _ := request(http.MethodGet, "/", e)
	assert.Equal(t, http.StatusInternalServerError, c)
}

func TestEchoHandler(t *testing.T) {
	e := New()

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.String(http.StatusOK, "OK")
		}
	})

	c, b := request(http.MethodGet, "/ok", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoWrapHandler(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test", rec.Body.String())
	}
}

func TestEchoWrapMiddleware(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	buf := new(bytes.Buffer)
	mw := WrapMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf.Write([]byte("mw"))
			h.ServeHTTP(w, r)
		})
	})
	h := mw(func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, "mw", buf.String())
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	}
}

func TestEchoContext(t *testing.T) {
	e := New()
	c := e.AcquireContext()
	assert.IsType(t, new(context), c)
	e.ReleaseContext(c)
}

func waitForServerStart(e *Echo, errChan <-chan error, isTLS bool) error {
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 200*time.Millisecond)
	defer cancel()

	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var addr net.Addr
			if isTLS {
				addr = e.TLSListenerAddr()
			} else {
				addr = e.ListenerAddr()
			}
			if addr != nil && strings.Contains(addr.String(), ":") {
				return nil // was started
			}
		case err := <-errChan:
			if err == http.ErrServerClosed {
				return nil
			}
			return err
		}
	}
}

func TestEchoStart(t *testing.T) {
	e := New()
	errChan := make(chan error)

	go func() {
		err := e.Start(":0")
		if err != nil {
			errChan <- err
		}
	}()

	err := waitForServerStart(e, errChan, false)
	assert.NoError(t, err)

	assert.NoError(t, e.Close())
}

func TestEcho_StartTLS(t *testing.T) {
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
			errChan := make(chan error)

			go func() {
				certFile := "_fixture/certs/cert.pem"
				if tc.certFile != "" {
					certFile = tc.certFile
				}
				keyFile := "_fixture/certs/key.pem"
				if tc.keyFile != "" {
					keyFile = tc.keyFile
				}

				err := e.StartTLS(tc.addr, certFile, keyFile)
				if err != nil {
					errChan <- err
				}
			}()

			err := waitForServerStart(e, errChan, true)
			if tc.expectError != "" {
				if _, ok := err.(*os.PathError); ok {
					assert.Error(t, err) // error messages for unix and windows are different. so test only error type here
				} else {
					assert.EqualError(t, err, tc.expectError)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, e.Close())
		})
	}
}

func TestEchoStartTLSAndStart(t *testing.T) {
	// We test if Echo and listeners work correctly when Echo is simultaneously attached to HTTP and HTTPS server
	e := New()

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.String(http.StatusOK, "OK")
		}
	})

	errTLSChan := make(chan error)
	go func() {
		certFile := "_fixture/certs/cert.pem"
		keyFile := "_fixture/certs/key.pem"
		err := e.StartTLS("localhost:", certFile, keyFile)
		if err != nil {
			errTLSChan <- err
		}
	}()

	err := waitForServerStart(e, errTLSChan, true)
	assert.NoError(t, err)
	defer func() {
		if err := e.Shutdown(stdContext.Background()); err != nil {
			t.Error(err)
		}
	}()

	// check if HTTPS works (note: we are using self signed certs so InsecureSkipVerify=true)
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	res, err := client.Get("https://" + e.TLSListenerAddr().String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	errChan := make(chan error)
	go func() {
		err := e.Start("localhost:")
		if err != nil {
			errChan <- err
		}
	}()
	err = waitForServerStart(e, errChan, false)
	assert.NoError(t, err)

	// now we are serving both HTTPS and HTTP listeners. see if HTTP works in addition to HTTPS
	res, err = http.Get("http://" + e.ListenerAddr().String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// see if HTTPS works after HTTP listener is also added
	res, err = client.Get("https://" + e.TLSListenerAddr().String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestEchoStartTLSByteString(t *testing.T) {
	cert, err := ioutil.ReadFile("_fixture/certs/cert.pem")
	require.NoError(t, err)
	key, err := ioutil.ReadFile("_fixture/certs/key.pem")
	require.NoError(t, err)

	testCases := []struct {
		cert        interface{}
		key         interface{}
		expectedErr error
		name        string
	}{
		{
			cert:        "_fixture/certs/cert.pem",
			key:         "_fixture/certs/key.pem",
			expectedErr: nil,
			name:        `ValidCertAndKeyFilePath`,
		},
		{
			cert:        cert,
			key:         key,
			expectedErr: nil,
			name:        `ValidCertAndKeyByteString`,
		},
		{
			cert:        cert,
			key:         1,
			expectedErr: ErrInvalidCertOrKeyType,
			name:        `InvalidKeyType`,
		},
		{
			cert:        0,
			key:         key,
			expectedErr: ErrInvalidCertOrKeyType,
			name:        `InvalidCertType`,
		},
		{
			cert:        0,
			key:         1,
			expectedErr: ErrInvalidCertOrKeyType,
			name:        `InvalidCertAndKeyTypes`,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			e := New()
			e.HideBanner = true

			errChan := make(chan error)

			go func() {
				errChan <- e.StartTLS(":0", test.cert, test.key)
			}()

			err := waitForServerStart(e, errChan, true)
			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, e.Close())
		})
	}
}

func TestEcho_StartAutoTLS(t *testing.T) {
	var testCases = []struct {
		name        string
		addr        string
		expectError string
	}{
		{
			name: "ok",
			addr: ":0",
		},
		{
			name:        "nok, invalid address",
			addr:        "nope",
			expectError: "listen tcp: address nope: missing port in address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			errChan := make(chan error)

			go func() {
				errChan <- e.StartAutoTLS(tc.addr)
			}()

			err := waitForServerStart(e, errChan, true)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, e.Close())
		})
	}
}

func TestEcho_StartH2CServer(t *testing.T) {
	var testCases = []struct {
		name        string
		addr        string
		expectError string
	}{
		{
			name: "ok",
			addr: ":0",
		},
		{
			name:        "nok, invalid address",
			addr:        "nope",
			expectError: "listen tcp: address nope: missing port in address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.Debug = true
			h2s := &http2.Server{}

			errChan := make(chan error)
			go func() {
				err := e.StartH2CServer(tc.addr, h2s)
				if err != nil {
					errChan <- err
				}
			}()

			err := waitForServerStart(e, errChan, false)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, e.Close())
		})
	}
}

func request(method, path string, e *Echo) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func TestHTTPError(t *testing.T) {
	t.Run("non-internal", func(t *testing.T) {
		err := NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"code": 12,
		})

		assert.Equal(t, "code=400, message=map[code:12]", err.Error())
	})
	t.Run("internal", func(t *testing.T) {
		err := NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"code": 12,
		})
		err.SetInternal(errors.New("internal error"))
		assert.Equal(t, "code=400, message=map[code:12], internal=internal error", err.Error())
	})
}

func TestHTTPError_Unwrap(t *testing.T) {
	t.Run("non-internal", func(t *testing.T) {
		err := NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"code": 12,
		})

		assert.Nil(t, errors.Unwrap(err))
	})
	t.Run("internal", func(t *testing.T) {
		err := NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"code": 12,
		})
		err.SetInternal(errors.New("internal error"))
		assert.Equal(t, "internal error", errors.Unwrap(err).Error())
	})
}

func TestDefaultHTTPErrorHandler(t *testing.T) {
	e := New()
	e.Debug = true

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			switch c.Request().URL.Path {
			case "/plain":
				return errors.New("An error occurred")
			case "/badrequest":
				return NewHTTPError(http.StatusBadRequest, "Invalid request")
			case "/servererror":
				return NewHTTPError(http.StatusInternalServerError, map[string]interface{}{
					"code":    33,
					"message": "Something bad happened",
					"error":   "stackinfo",
				})
			case "/early-return":
				c.String(http.StatusOK, "OK")
				return errors.New("ERROR")
			case "/internal-error":
				err := errors.New("internal error message body")
				return NewHTTPError(http.StatusBadRequest).SetInternal(err)
			default:
				return NewHTTPError(http.StatusNotFound)
			}
		}
	})

	// With Debug=true plain response contains error message
	c, b := request(http.MethodGet, "/plain", e)
	assert.Equal(t, http.StatusInternalServerError, c)
	assert.Equal(t, "{\n  \"error\": \"An error occurred\",\n  \"message\": \"Internal Server Error\"\n}\n", b)
	// and special handling for HTTPError
	c, b = request(http.MethodGet, "/badrequest", e)
	assert.Equal(t, http.StatusBadRequest, c)
	assert.Equal(t, "{\n  \"error\": \"code=400, message=Invalid request\",\n  \"message\": \"Invalid request\"\n}\n", b)
	// complex errors are serialized to pretty JSON
	c, b = request(http.MethodGet, "/servererror", e)
	assert.Equal(t, http.StatusInternalServerError, c)
	assert.Equal(t, "{\n  \"code\": 33,\n  \"error\": \"stackinfo\",\n  \"message\": \"Something bad happened\"\n}\n", b)
	// if the body is already set HTTPErrorHandler should not add anything to response body
	c, b = request(http.MethodGet, "/early-return", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
	// internal error should be reflected in the message
	c, b = request(http.MethodGet, "/internal-error", e)
	assert.Equal(t, http.StatusBadRequest, c)
	assert.Equal(t, "{\n  \"error\": \"code=400, message=Bad Request, internal=internal error message body\",\n  \"message\": \"Bad Request\"\n}\n", b)

	e.Debug = false
	// With Debug=false the error response is shortened
	c, b = request(http.MethodGet, "/plain", e)
	assert.Equal(t, http.StatusInternalServerError, c)
	assert.Equal(t, "{\"message\":\"Internal Server Error\"}\n", b)
	c, b = request(http.MethodGet, "/badrequest", e)
	assert.Equal(t, http.StatusBadRequest, c)
	assert.Equal(t, "{\"message\":\"Invalid request\"}\n", b)
	// No difference for error response with non plain string errors
	c, b = request(http.MethodGet, "/servererror", e)
	assert.Equal(t, http.StatusInternalServerError, c)
	assert.Equal(t, "{\"code\":33,\"error\":\"stackinfo\",\"message\":\"Something bad happened\"}\n", b)
}

func TestEchoClose(t *testing.T) {
	e := New()
	errCh := make(chan error)

	go func() {
		errCh <- e.Start(":0")
	}()

	err := waitForServerStart(e, errCh, false)
	assert.NoError(t, err)

	if err := e.Close(); err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, e.Close())

	err = <-errCh
	assert.Equal(t, err.Error(), "http: Server closed")
}

func TestEchoShutdown(t *testing.T) {
	e := New()
	errCh := make(chan error)

	go func() {
		errCh <- e.Start(":0")
	}()

	err := waitForServerStart(e, errCh, false)
	assert.NoError(t, err)

	if err := e.Close(); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 10*time.Second)
	defer cancel()
	assert.NoError(t, e.Shutdown(ctx))

	err = <-errCh
	assert.Equal(t, err.Error(), "http: Server closed")
}

var listenerNetworkTests = []struct {
	test    string
	network string
	address string
}{
	{"tcp ipv4 address", "tcp", "127.0.0.1:1323"},
	{"tcp ipv6 address", "tcp", "[::1]:1323"},
	{"tcp4 ipv4 address", "tcp4", "127.0.0.1:1323"},
	{"tcp6 ipv6 address", "tcp6", "[::1]:1323"},
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

func TestEchoListenerNetwork(t *testing.T) {
	hasIPv6 := supportsIPv6()
	for _, tt := range listenerNetworkTests {
		if !hasIPv6 && strings.Contains(tt.address, "::") {
			t.Skip("Skipping testing IPv6 for " + tt.address + ", not available")
			continue
		}
		t.Run(tt.test, func(t *testing.T) {
			e := New()
			e.ListenerNetwork = tt.network

			// HandlerFunc
			e.Use(func(next HandlerFunc) HandlerFunc {
				return func(c Context) error {
					return c.String(http.StatusOK, "OK")
				}
			})

			errCh := make(chan error)

			go func() {
				errCh <- e.Start(tt.address)
			}()

			err := waitForServerStart(e, errCh, false)
			assert.NoError(t, err)

			if resp, err := http.Get(fmt.Sprintf("http://%s/ok", tt.address)); err == nil {
				defer resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				if body, err := ioutil.ReadAll(resp.Body); err == nil {
					assert.Equal(t, "OK", string(body))
				} else {
					assert.Fail(t, err.Error())
				}

			} else {
				assert.Fail(t, err.Error())
			}

			if err := e.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEchoListenerNetworkInvalid(t *testing.T) {
	e := New()
	e.ListenerNetwork = "unix"

	// HandlerFunc
	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.String(http.StatusOK, "OK")
		}
	})

	assert.Equal(t, ErrInvalidListenerNetwork, e.Start(":1323"))
}

func TestEcho_ListenerAddr(t *testing.T) {
	e := New()

	addr := e.ListenerAddr()
	assert.Nil(t, addr)

	errCh := make(chan error)
	go func() {
		errCh <- e.Start(":0")
	}()

	err := waitForServerStart(e, errCh, false)
	assert.NoError(t, err)
}

func TestEcho_TLSListenerAddr(t *testing.T) {
	cert, err := ioutil.ReadFile("_fixture/certs/cert.pem")
	require.NoError(t, err)
	key, err := ioutil.ReadFile("_fixture/certs/key.pem")
	require.NoError(t, err)

	e := New()

	addr := e.TLSListenerAddr()
	assert.Nil(t, addr)

	errCh := make(chan error)
	go func() {
		errCh <- e.StartTLS(":0", cert, key)
	}()

	err = waitForServerStart(e, errCh, true)
	assert.NoError(t, err)
}

func TestEcho_StartServer(t *testing.T) {
	cert, err := ioutil.ReadFile("_fixture/certs/cert.pem")
	require.NoError(t, err)
	key, err := ioutil.ReadFile("_fixture/certs/key.pem")
	require.NoError(t, err)
	certs, err := tls.X509KeyPair(cert, key)
	require.NoError(t, err)

	var testCases = []struct {
		name        string
		addr        string
		TLSConfig   *tls.Config
		expectError string
	}{
		{
			name: "ok",
			addr: ":0",
		},
		{
			name:      "ok, start with TLS",
			addr:      ":0",
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{certs}},
		},
		{
			name:        "nok, invalid address",
			addr:        "nope",
			expectError: "listen tcp: address nope: missing port in address",
		},
		{
			name:        "nok, invalid tls address",
			addr:        "nope",
			TLSConfig:   &tls.Config{InsecureSkipVerify: true},
			expectError: "listen tcp: address nope: missing port in address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.Debug = true

			server := new(http.Server)
			server.Addr = tc.addr
			if tc.TLSConfig != nil {
				server.TLSConfig = tc.TLSConfig
			}

			errCh := make(chan error)
			go func() {
				errCh <- e.StartServer(server)
			}()

			err := waitForServerStart(e, errCh, tc.TLSConfig != nil)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, e.Close())
		})
	}
}
