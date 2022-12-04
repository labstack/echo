package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Note: Just for the test coverage, not a real test.
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
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
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = Logger()(func(c echo.Context) error {
		return errors.New("error")
	})
	h(c)
}

func TestLoggerIPAddress(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
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
	assert.Contains(t, buf.String(), ip)

	// With X-Forwarded-For
	buf.Reset()
	req.Header.Del(echo.HeaderXRealIP)
	req.Header.Add(echo.HeaderXForwardedFor, ip)
	h(c)
	assert.Contains(t, buf.String(), ip)

	buf.Reset()
	h(c)
	assert.Contains(t, buf.String(), ip)
}

func TestLoggerTemplate(t *testing.T) {
	buf := new(bytes.Buffer)

	e := echo.New()
	e.Use(LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "route":"${route}", "referer":"${referer}",` +
			`"bytes_out":${bytes_out},"ch":"${header:X-Custom-Header}", "protocol":"${protocol}"` +
			`"us":"${query:username}", "cf":"${form:username}", "session":"${cookie:session}"}` + "\n",
		Output: buf,
	}))

	e.GET("/users/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, "Header Logged")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/1?username=apagano-param&password=secret", nil)
	req.RequestURI = "/"
	req.Header.Add(echo.HeaderXRealIP, "127.0.0.1")
	req.Header.Add("Referer", "google.com")
	req.Header.Add("User-Agent", "echo-tests-agent")
	req.Header.Add("X-Custom-Header", "AAA-CUSTOM-VALUE")
	req.Header.Add("X-Request-ID", "6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	req.Header.Add("Cookie", "_ga=GA1.2.000000000.0000000000; session=ac08034cd216a647fc2eb62f2bcf7b810")
	req.Form = url.Values{
		"username": []string{"apagano-form"},
		"password": []string{"secret-form"},
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	cases := map[string]bool{
		"apagano-param":                        true,
		"apagano-form":                         true,
		"AAA-CUSTOM-VALUE":                     true,
		"BBB-CUSTOM-VALUE":                     false,
		"secret-form":                          false,
		"hexvalue":                             false,
		"GET":                                  true,
		"127.0.0.1":                            true,
		"\"path\":\"/users/1\"":                true,
		"\"route\":\"/users/:id\"":             true,
		"\"uri\":\"/\"":                        true,
		"\"status\":200":                       true,
		"\"bytes_in\":0":                       true,
		"google.com":                           true,
		"echo-tests-agent":                     true,
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8": true,
		"ac08034cd216a647fc2eb62f2bcf7b810":    true,
	}

	for token, present := range cases {
		assert.True(t, strings.Contains(buf.String(), token) == present, "Case: "+token)
	}
}

func TestLoggerCustomTimestamp(t *testing.T) {
	buf := new(bytes.Buffer)
	customTimeFormat := "2006-01-02 15:04:05.00000"
	e := echo.New()
	e.Use(LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time_custom}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "referer":"${referer}",` +
			`"bytes_out":${bytes_out},"ch":"${header:X-Custom-Header}",` +
			`"us":"${query:username}", "cf":"${form:username}", "session":"${cookie:session}"}` + "\n",
		CustomTimeFormat: customTimeFormat,
		Output:           buf,
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "custom time stamp test")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var objs map[string]*json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &objs); err != nil {
		panic(err)
	}
	loggedTime := *(*string)(unsafe.Pointer(objs["time"]))
	_, err := time.Parse(customTimeFormat, loggedTime)
	assert.Error(t, err)
}

func TestLoggerCustomTagFunc(t *testing.T) {
	e := echo.New()
	buf := new(bytes.Buffer)
	e.Use(LoggerWithConfig(LoggerConfig{
		Format: `{"method":"${method}",${custom}}` + "\n",
		CustomTagFunc: func(c echo.Context, buf *bytes.Buffer) (int, error) {
			return buf.WriteString(`"tag":"my-value"`)
		},
		Output: buf,
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "custom time stamp test")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, `{"method":"GET","tag":"my-value"}`+"\n", buf.String())
}

func BenchmarkLoggerWithConfig_withoutMapFields(b *testing.B) {
	e := echo.New()

	buf := new(bytes.Buffer)
	mw := LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "referer":"${referer}",` +
			`"bytes_out":${bytes_out}, "protocol":"${protocol}"}` + "\n",
		Output: buf,
	})(func(c echo.Context) error {
		c.Request().Header.Set(echo.HeaderXRequestID, "123")
		c.FormValue("to force parse form")
		return c.String(http.StatusTeapot, "OK")
	})

	f := make(url.Values)
	f.Set("csrf", "token")
	f.Add("multiple", "1")
	f.Add("multiple", "2")
	req := httptest.NewRequest(http.MethodPost, "/test?lang=en&checked=1&checked=2", strings.NewReader(f.Encode()))
	req.Header.Set("Referer", "https://echo.labstack.com/")
	req.Header.Set("User-Agent", "curl/7.68.0")
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		mw(c)
		buf.Reset()
	}
}

func BenchmarkLoggerWithConfig_withMapFields(b *testing.B) {
	e := echo.New()

	buf := new(bytes.Buffer)
	mw := LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "referer":"${referer}",` +
			`"bytes_out":${bytes_out},"ch":"${header:X-Custom-Header}", "protocol":"${protocol}"` +
			`"us":"${query:username}", "cf":"${form:csrf}", "Referer2":"${header:Referer}"}` + "\n",
		Output: buf,
	})(func(c echo.Context) error {
		c.Request().Header.Set(echo.HeaderXRequestID, "123")
		c.FormValue("to force parse form")
		return c.String(http.StatusTeapot, "OK")
	})

	f := make(url.Values)
	f.Set("csrf", "token")
	f.Add("multiple", "1")
	f.Add("multiple", "2")
	req := httptest.NewRequest(http.MethodPost, "/test?lang=en&checked=1&checked=2", strings.NewReader(f.Encode()))
	req.Header.Set("Referer", "https://echo.labstack.com/")
	req.Header.Set("User-Agent", "curl/7.68.0")
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		mw(c)
		buf.Reset()
	}
}

func TestLoggerTemplateWithTimeUnixMilli(t *testing.T) {
	buf := new(bytes.Buffer)

	e := echo.New()
	e.Use(LoggerWithConfig(LoggerConfig{
		Format: `${time_unix_milli}`,
		Output: buf,
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	unixMillis, err := strconv.ParseInt(buf.String(), 10, 64)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Unix(unixMillis/1000, 0), time.Now(), 3*time.Second)
}

func TestLoggerTemplateWithTimeUnixMicro(t *testing.T) {
	buf := new(bytes.Buffer)

	e := echo.New()
	e.Use(LoggerWithConfig(LoggerConfig{
		Format: `${time_unix_micro}`,
		Output: buf,
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	unixMicros, err := strconv.ParseInt(buf.String(), 10, 64)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Unix(unixMicros/1000000, 0), time.Now(), 3*time.Second)
}
