package middleware

import (
	"bytes"
	"github.com/andybalholm/brotli"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBrotli(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Skip if no Accept-Encoding header
	h := Brotli()(func(c echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})
	h(c)

	assert.Equal(t, "test", rec.Body.String())

	// Brotli
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, brotliScheme)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h(c)
	assert.Equal(t, brotliScheme, rec.Header().Get(echo.HeaderContentEncoding))
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), echo.MIMETextPlain)
	r := brotli.NewReader(rec.Body)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	assert.Equal(t, "test", buf.String())

	chunkBuf := make([]byte, 5)

	// Brotli chunked
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, brotliScheme)
	rec = httptest.NewRecorder()

	c = e.NewContext(req, rec)
	Brotli()(func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Transfer-Encoding", "chunked")

		// Write and flush the first part of the data
		c.Response().Write([]byte("test\n"))
		c.Response().Flush()

		// Read the first part of the data
		assert.True(t, rec.Flushed)
		assert.Equal(t, brotliScheme, rec.Header().Get(echo.HeaderContentEncoding))
		r.Reset(rec.Body)

		_, err := io.ReadFull(r, chunkBuf)
		assert.NoError(t, err)
		assert.Equal(t, "test\n", string(chunkBuf))

		// Write and flush the second part of the data
		c.Response().Write([]byte("test\n"))
		c.Response().Flush()

		_, err = io.ReadFull(r, chunkBuf)
		assert.NoError(t, err)
		assert.Equal(t, "test\n", string(chunkBuf))

		// Write the final part of the data and return
		c.Response().Write([]byte("test"))
		return nil
	})(c)

	buf = new(bytes.Buffer)
	buf.ReadFrom(r)
	assert.Equal(t, "test", buf.String())
}

func TestBrotliNoContent(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, brotliScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Brotli()(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	if assert.NoError(t, h(c)) {
		assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func TestBrotliEmpty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, brotliScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Brotli()(func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, brotliScheme, rec.Header().Get(echo.HeaderContentEncoding))
		assert.Equal(t, "text/plain; charset=UTF-8", rec.Header().Get(echo.HeaderContentType))
		r := brotli.NewReader(rec.Body)
		var buf bytes.Buffer
		buf.ReadFrom(r)
		assert.Equal(t, "", buf.String())
	}
}

func TestBrotliErrorReturned(t *testing.T) {
	e := echo.New()
	e.Use(Brotli())
	e.GET("/", func(c echo.Context) error {
		return echo.ErrNotFound
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, brotliScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
}

// Issue #806
func TestBrotliWithStatic(t *testing.T) {
	e := echo.New()
	e.Use(Brotli())
	e.Static("/test", "../_fixture/images")
	req := httptest.NewRequest(http.MethodGet, "/test/walle.png", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, brotliScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	// Data is written out in chunks when Content-Length == "", so only
	// validate the content length if it's not set.
	if cl := rec.Header().Get("Content-Length"); cl != "" {
		assert.Equal(t, cl, rec.Body.Len())
	}
	r := brotli.NewReader(rec.Body)
	want, err := ioutil.ReadFile("../_fixture/images/walle.png")
	if assert.NoError(t, err) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r)
		assert.Equal(t, want, buf.Bytes())
	}
}

func BenchmarkBrotli(b *testing.B) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, brotliScheme)

	h := Brotli()(func(c echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Brotli
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h(c)
	}
}
