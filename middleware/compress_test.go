package middleware

import (
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestGzip(t *testing.T) {
	// Empty Accept-Encoding header
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec), echo.New())
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	Gzip()(h)(c)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test", rec.Body.String())

	// With Accept-Encoding header
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
