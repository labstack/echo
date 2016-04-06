package middleware

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestGzip(t *testing.T) {
	e := echo.New()
	rq := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := echo.NewContext(rq, rec, e)

	// Skip if no Accept-Encoding header
	h := Gzip()(func(c echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})
	h(c)
	assert.Equal(t, "test", rec.Body.String())

	rq = test.NewRequest(echo.GET, "/", nil)
	rq.Header().Set(echo.HeaderAcceptEncoding, "gzip")
	rec = test.NewResponseRecorder()
	c = echo.NewContext(rq, rec, e)

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
	rq := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := echo.NewContext(rq, rec, e)
	h := Gzip()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	h(c)

	assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
	assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
	b, err := ioutil.ReadAll(rec.Body)
	if assert.NoError(t, err) {
		assert.Equal(t, 0, len(b))
	}
}

func TestGzipErrorReturned(t *testing.T) {
	e := echo.New()
	e.Use(Gzip())
	e.Get("/", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "error")
	})
	rq := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	e.ServeHTTP(rq, rec)

	assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
	b, err := ioutil.ReadAll(rec.Body)
	if assert.NoError(t, err) {
		assert.Equal(t, "error", string(b))
	}
}
