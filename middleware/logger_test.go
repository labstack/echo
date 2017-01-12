package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Note: Just for the test coverage, not a real test.
	e := echo.New()
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Logger()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Status 2xx
	h(c)

	// Status 3xx
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = Logger()(func(c echo.Context) error {
		return c.String(http.StatusTemporaryRedirect, "test")
	})
	h(c)

	// Status 4xx
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = Logger()(func(c echo.Context) error {
		return c.String(http.StatusNotFound, "test")
	})
	h(c)

	// Status 5xx with empty path
	req, _ = http.NewRequest(echo.GET, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = Logger()(func(c echo.Context) error {
		return errors.New("error")
	})
	h(c)
}

func TestLoggerIPAddress(t *testing.T) {
	e := echo.New()
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	buf := new(bytes.Buffer)
	e.Logger.SetOutput(buf)
	ip := "127.0.0.1"
	h := Logger()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// With X-Real-IP
	req.Header.Add(echo.HeaderXRealIP, ip)
	h(c)
	assert.Contains(t, ip, buf.String())

	// With X-Forwarded-For
	buf.Reset()
	req.Header.Del(echo.HeaderXRealIP)
	req.Header.Add(echo.HeaderXForwardedFor, ip)
	h(c)
	assert.Contains(t, ip, buf.String())

	buf.Reset()
	h(c)
	assert.Contains(t, ip, buf.String())
}

func TestLoggerFields(t *testing.T) {
	buf := new(bytes.Buffer)

	e := echo.New()
	e.Use(LoggerWithConfig(LoggerConfig{
		Fields: []string{
			"time",
			"remote_ip",
			"host",
			"user_agent",
			"method",
			"uri",
			"path",
			"referer",
			"status",
			"latency",
			"latency_human",
			"bytes_in",
			"bytes_out",
			"header:X-Custom-Header",
			"query:user",
			"form:user",
		},
		Output: &Stream{buf},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest(echo.GET, "/?user=joe", nil)
	req.RequestURI = "/"
	req.Header.Add(echo.HeaderXRealIP, "127.0.0.1")
	req.Header.Add("Referer", "google.com")
	req.Header.Add("User-Agent", "test-agent")
	req.Header.Add("X-Custom-Header", "AAA-CUSTOM-VALUE")
	req.Form = url.Values{
		"user": []string{"jon"},
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	cases := map[string]bool{
		"time":             true,
		"joe":              true,
		"jon":              true,
		"AAA-CUSTOM-VALUE": true,
		"GET":              true,
		"127.0.0.1":        true,
		"\"path\":\"/\"":   true,
		"\"uri\":\"/\"":    true,
		"\"status\":200":   true,
		"\"bytes_in\":0":   true,
		"google.com":       true,
		"test-agent":       true,
	}

	for token, present := range cases {
		assert.True(t, strings.Contains(buf.String(), token) == present, "Case: "+token)
	}
}
