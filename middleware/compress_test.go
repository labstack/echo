package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestGzip_NoAcceptEncodingHeader(t *testing.T) {
	// Skip if no Accept-Encoding header
	h := Gzip()(func(c echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	assert.NoError(t, err)

	assert.Equal(t, "test", rec.Body.String())
}

func TestMustGzipWithConfig_panics(t *testing.T) {
	assert.Panics(t, func() {
		GzipWithConfig(GzipConfig{Level: 999})
	})
}

func TestGzip_AcceptEncodingHeader(t *testing.T) {
	h := Gzip()(func(c echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	assert.NoError(t, err)

	assert.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), echo.MIMETextPlain)

	r, err := gzip.NewReader(rec.Body)
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	defer r.Close()
	buf.ReadFrom(r)
	assert.Equal(t, "test", buf.String())
}

func TestGzip_chunked(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	chunkChan := make(chan struct{})
	waitChan := make(chan struct{})
	h := Gzip()(func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Transfer-Encoding", "chunked")

		// Write and flush the first part of the data
		c.Response().Write([]byte("first\n"))
		c.Response().Flush()

		chunkChan <- struct{}{}
		<-waitChan

		// Write and flush the second part of the data
		c.Response().Write([]byte("second\n"))
		c.Response().Flush()

		chunkChan <- struct{}{}
		<-waitChan

		// Write the final part of the data and return
		c.Response().Write([]byte("third"))

		chunkChan <- struct{}{}
		return nil
	})

	go func() {
		err := h(c)
		chunkChan <- struct{}{}
		assert.NoError(t, err)
	}()

	<-chunkChan // wait for first write
	waitChan <- struct{}{}

	<-chunkChan // wait for second write
	waitChan <- struct{}{}

	<-chunkChan                      // wait for final write in handler
	<-chunkChan                      // wait for return from handler
	time.Sleep(5 * time.Millisecond) // to have time for flushing

	assert.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))

	r, err := gzip.NewReader(rec.Body)
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	assert.Equal(t, "first\nsecond\nthird", buf.String())
}

func TestGzip_NoContent(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Gzip()(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	if assert.NoError(t, h(c)) {
		assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func TestGzip_Empty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Gzip()(func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
		assert.Equal(t, "text/plain; charset=UTF-8", rec.Header().Get(echo.HeaderContentType))
		r, err := gzip.NewReader(rec.Body)
		if assert.NoError(t, err) {
			var buf bytes.Buffer
			buf.ReadFrom(r)
			assert.Equal(t, "", buf.String())
		}
	}
}

func TestGzip_ErrorReturned(t *testing.T) {
	e := echo.New()
	e.Use(Gzip())
	e.GET("/", func(c echo.Context) error {
		return echo.ErrNotFound
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
}

func TestGzipWithConfig_invalidLevel(t *testing.T) {
	mw, err := GzipConfig{Level: 12}.ToMiddleware()
	assert.EqualError(t, err, "invalid gzip level")
	assert.Nil(t, mw)
}

// Issue #806
func TestGzipWithStatic(t *testing.T) {
	e := echo.New()
	e.Filesystem = os.DirFS("../")

	e.Use(Gzip())
	e.Static("/test", "_fixture/images")
	req := httptest.NewRequest(http.MethodGet, "/test/walle.png", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Data is written out in chunks when Content-Length == "", so only
	// validate the content length if it's not set.
	if cl := rec.Header().Get("Content-Length"); cl != "" {
		assert.Equal(t, cl, rec.Body.Len())
	}
	r, err := gzip.NewReader(rec.Body)
	if assert.NoError(t, err) {
		defer r.Close()
		want, err := os.ReadFile("../_fixture/images/walle.png")
		if assert.NoError(t, err) {
			buf := new(bytes.Buffer)
			buf.ReadFrom(r)
			assert.Equal(t, want, buf.Bytes())
		}
	}
}

func TestGzipWithMinLength(t *testing.T) {
	e := echo.New()
	// Minimal response length
	e.Use(GzipWithConfig(GzipConfig{MinLength: 10}))
	e.GET("/", func(c echo.Context) error {
		c.Response().Write([]byte("foobarfoobar"))
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
	r, err := gzip.NewReader(rec.Body)
	if assert.NoError(t, err) {
		buf := new(bytes.Buffer)
		defer r.Close()
		buf.ReadFrom(r)
		assert.Equal(t, "foobarfoobar", buf.String())
	}
}

func TestGzipWithMinLengthTooShort(t *testing.T) {
	e := echo.New()
	// Minimal response length
	e.Use(GzipWithConfig(GzipConfig{MinLength: 10}))
	e.GET("/", func(c echo.Context) error {
		c.Response().Write([]byte("test"))
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "", rec.Header().Get(echo.HeaderContentEncoding))
	assert.Contains(t, rec.Body.String(), "test")
}

func TestGzipWithResponseWithoutBody(t *testing.T) {
	e := echo.New()

	e.Use(Gzip())
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "http://localhost")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "", rec.Header().Get(echo.HeaderContentEncoding))
}

func TestGzipWithMinLengthChunked(t *testing.T) {
	e := echo.New()

	// Gzip chunked
	chunkBuf := make([]byte, 5)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()

	var r *gzip.Reader = nil

	c := e.NewContext(req, rec)
	next := func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Transfer-Encoding", "chunked")

		// Write and flush the first part of the data
		c.Response().Write([]byte("test\n"))
		c.Response().Flush()

		// Read the first part of the data
		assert.True(t, rec.Flushed)
		assert.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))

		var err error
		r, err = gzip.NewReader(rec.Body)
		assert.NoError(t, err)

		_, err = io.ReadFull(r, chunkBuf)
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
	}
	err := GzipWithConfig(GzipConfig{MinLength: 10})(next)(c)

	assert.NoError(t, err)
	assert.NotNil(t, r)

	buf := new(bytes.Buffer)

	buf.ReadFrom(r)
	assert.Equal(t, "test", buf.String())

	r.Close()
}

func TestGzipWithMinLengthNoContent(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := GzipWithConfig(GzipConfig{MinLength: 10})(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	if assert.NoError(t, h(c)) {
		assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func BenchmarkGzip(b *testing.B) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)

	h := Gzip()(func(c echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Gzip
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h(c)
	}
}
