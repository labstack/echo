package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestGzipWriteReturnsCorrectCount verifies that gzipResponseWriter.Write honours
// the io.Writer contract: it must return n == len(b) (never more) for the slice
// passed to a single Write call. Before the fix, when a buffered write crossed the
// MinLength threshold, Write returned the full buffer length (previous chunks + b),
// over-reporting the count and panicking callers like io.Copy.
func TestGzipWriteReturnsCorrectCount(t *testing.T) {
	e := echo.New()
	mw := GzipWithConfig(GzipConfig{MinLength: 100})

	var n1, n2 int
	h := mw(func(c echo.Context) error {
		chunk1 := []byte("hello ")               // 6 bytes, stays below MinLength
		chunk2 := bytes.Repeat([]byte("x"), 200) // crosses MinLength
		var err error
		if n1, err = c.Response().Write(chunk1); err != nil {
			return err
		}
		if n2, err = c.Response().Write(chunk2); err != nil {
			return err
		}
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	assert.NoError(t, h(e.NewContext(req, rec)))

	// Each Write must report exactly the length of the slice it was given.
	assert.Equal(t, 6, n1, "first Write should report len of its own slice")
	assert.Equal(t, 200, n2, "second Write must report len(b), not the buffered total")
}

// TestGzipIoCopyDoesNotPanic reproduces the real-world failure: streaming through
// the gzip response writer with io.Copy (as echo.Context#Stream does) panics with
// "invalid write count" when Write over-reports the byte count.
func TestGzipIoCopyDoesNotPanic(t *testing.T) {
	e := echo.New()
	mw := GzipWithConfig(GzipConfig{MinLength: 100})

	h := mw(func(c echo.Context) error {
		// Small write keeps us below MinLength so the buffer holds previous bytes.
		if _, err := c.Response().Write([]byte("prefix")); err != nil {
			return err
		}
		// io.Copy validates that the returned write count never exceeds len(p);
		// an over-reported count makes bytes.Reader.WriteTo panic.
		src := bytes.NewReader(bytes.Repeat([]byte("y"), 200))
		_, err := io.Copy(c.Response(), src)
		return err
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		assert.NoError(t, h(e.NewContext(req, rec)))
	})
}
