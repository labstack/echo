package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Note: Just for the test coverage, not a real test.
	e := echo.New()
	rq := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := echo.NewContext(rq, rec, e)
	h := Logger()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Status 2xx
	h(c)

	// Status 3xx
	rec = test.NewResponseRecorder()
	c = echo.NewContext(rq, rec, e)
	h = Logger()(func(c echo.Context) error {
		return c.String(http.StatusTemporaryRedirect, "test")
	})
	h(c)

	// Status 4xx
	rec = test.NewResponseRecorder()
	c = echo.NewContext(rq, rec, e)
	h = Logger()(func(c echo.Context) error {
		return c.String(http.StatusNotFound, "test")
	})
	h(c)

	// Status 5xx with empty path
	rq = test.NewRequest(echo.GET, "", nil)
	rec = test.NewResponseRecorder()
	c = echo.NewContext(rq, rec, e)
	h = Logger()(func(c echo.Context) error {
		return errors.New("error")
	})
	h(c)
}

func TestLoggerIPAddress(t *testing.T) {
	e := echo.New()
	rq := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := echo.NewContext(rq, rec, e)
	buf := new(bytes.Buffer)
	e.Logger().SetOutput(buf)
	ip := "127.0.0.1"
	h := Logger()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// With X-Real-IP
	rq.Header().Add(echo.HeaderXRealIP, ip)
	h(c)
	assert.Contains(t, ip, buf.String())

	// With X-Forwarded-For
	buf.Reset()
	rq.Header().Del(echo.HeaderXRealIP)
	rq.Header().Add(echo.HeaderXForwardedFor, ip)
	h(c)
	assert.Contains(t, ip, buf.String())

	// with rq.RemoteAddr
	buf.Reset()
	h(c)
	assert.Contains(t, ip, buf.String())
}
