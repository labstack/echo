package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestLoggerHeaders(t *testing.T) {
	e := echo.New()
	buf := new(bytes.Buffer)

	logger := LoggerWithConfig(LoggerConfig{
		Skipper: defaultSkipper,
		Format: `{"time":"${time_rfc3339}","remote_ip":"${remote_ip}","host":"${host}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
			`"bytes_out":${bytes_out},"custom_header":"${header:X-Custom-Header}"}` + "\n",
		Output: buf,
	})

	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := logger(func(c echo.Context) error {
		return c.String(http.StatusOK, "Header Logged")
	})

	req.Header.Add("X-Custom-Header", "AAA-CUSTOM-VALUE")
	req.Header.Add("X-Custom-B-Header", "BBB-CUSTOM-VALUE")
	h(c)

	assert.Contains(t, buf.String(), "AAA-CUSTOM-VALUE")
	assert.NotContains(t, buf.String(), "BBB-CUSTOM-VALUE")
}

func TestLoggerQuery(t *testing.T) {
	e := echo.New()
	buf := new(bytes.Buffer)

	logger := LoggerWithConfig(LoggerConfig{
		Skipper: defaultSkipper,
		Format: `{"time":"${time_rfc3339}","remote_ip":"${remote_ip}","host":"${host}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
			`"bytes_out":${bytes_out},"custom_header":"${query:username}"}` + "\n",
		Output: buf,
	})

	req, _ := http.NewRequest(echo.GET, "/?username=apagano&password=secret", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := logger(func(c echo.Context) error {
		return c.String(http.StatusOK, "Header Logged")
	})

	h(c)

	assert.Contains(t, buf.String(), "apagano")
	assert.NotContains(t, buf.String(), "secret")
}

func TestLoggerForm(t *testing.T) {
	e := echo.New()
	buf := new(bytes.Buffer)

	logger := LoggerWithConfig(LoggerConfig{
		Skipper: defaultSkipper,
		Format: `{"time":"${time_rfc3339}","remote_ip":"${remote_ip}","host":"${host}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
			`"bytes_out":${bytes_out},"custom_header":"${form:username}"}` + "\n",
		Output: buf,
	})

	req, _ := http.NewRequest(echo.POST, "/", nil)
	req.Form = url.Values{
		"username": []string{"apagano"},
		"password": []string{"secret"},
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := logger(func(c echo.Context) error {
		return c.String(http.StatusOK, "Header Logged")
	})

	h(c)

	assert.Contains(t, buf.String(), "apagano")
	assert.NotContains(t, buf.String(), "secret")
}

func TestLoggerPath(t *testing.T) {
	e := echo.New()
	buf := new(bytes.Buffer)

	logger := LoggerWithConfig(LoggerConfig{
		Skipper: defaultSkipper,
		Format: `{"time":"${time_rfc3339}","remote_ip":"${remote_ip}","host":"${host}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
			`"bytes_out":${bytes_out},"custom_header":"${path:username}"}` + "\n",
		Output: buf,
	})

	req, _ := http.NewRequest(echo.POST, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("username", "hash")
	c.SetParamValues("apagano", "hexvalue")

	h := logger(func(c echo.Context) error {
		return c.String(http.StatusOK, "Header Logged")
	})

	h(c)

	assert.Contains(t, buf.String(), "apagano")
	assert.NotContains(t, buf.String(), "hexvalue")
}
