package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestGzip(t *testing.T) {
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec), echo.New())
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Skip if no Accept-Encoding header
	Gzip()(h)(c)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test", rec.Body.String())

	// Gzip
	req, _ = http.NewRequest(echo.GET, "/", nil)
	req.Header.Set(echo.AcceptEncoding, "gzip")
	rec = httptest.NewRecorder()
	c = echo.NewContext(req, echo.NewResponse(rec), echo.New())
	Gzip()(h)(c)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get(echo.ContentEncoding))
	r, err := gzip.NewReader(rec.Body)
	defer r.Close()
	if assert.NoError(t, err) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r)
		assert.Equal(t, "test", buf.String())
	}
}
