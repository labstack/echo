package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestGzip(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)

	// Skip if no Accept-Encoding header
	h := Gzip()(func(c echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})
	h(c)
	assert.Equal(t, "test", rec.Body.String())

	req = test.NewRequest(echo.GET, "/", nil)
	req.Header().Set(echo.HeaderAcceptEncoding, "gzip")
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)

	// Gzip
	h(c)
	assert.Equal(t, "gzip", rec.Header().Get(echo.HeaderContentEncoding))
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), echo.MIMETextPlain)
	r, err := gzip.NewReader(rec.Body)
	defer r.Close()
	if assert.NoError(t, err) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r)
		assert.Equal(t, "test", buf.String())
	}
}

func TestGzipNoContent(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	h := Gzip()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	if assert.NoError(t, h(c)) {
		assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func TestGzipErrorReturned(t *testing.T) {
	e := echo.New()
	e.Use(Gzip())
	e.GET("/", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "error")
	})
	req := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	e.ServeHTTP(req, rec)
	assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
	assert.Equal(t, "error", rec.Body.String())
}
