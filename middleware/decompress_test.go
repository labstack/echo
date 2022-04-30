package middleware

import (
	"bytes"
	"compress/gzip"
	"errors"
	"github.com/siyual-park/echo-slim/v4"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestDecompress(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Skip if no Content-Encoding header
	h := Decompress()(func(c echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})
	h(c)

	assert := assert.New(t)
	assert.Equal("test", rec.Body.String())

	// Decompress
	body := `{"name": "echo"}`
	gz, _ := gzipString(body)
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(gz)))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h(c)
	assert.Equal(GZIPEncoding, req.Header.Get(echo.HeaderContentEncoding))
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)
	assert.Equal(body, string(b))
}

func TestDecompressDefaultConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := DecompressWithConfig(DecompressConfig{})(func(c echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})
	h(c)

	assert := assert.New(t)
	assert.Equal("test", rec.Body.String())

	// Decompress
	body := `{"name": "echo"}`
	gz, _ := gzipString(body)
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(gz)))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h(c)
	assert.Equal(GZIPEncoding, req.Header.Get(echo.HeaderContentEncoding))
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)
	assert.Equal(body, string(b))
}

func TestCompressRequestWithoutDecompressMiddleware(t *testing.T) {
	e := echo.New()
	body := `{"name":"echo"}`
	gz, _ := gzipString(body)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(gz)))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	assert.Equal(t, GZIPEncoding, req.Header.Get(echo.HeaderContentEncoding))
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.NotEqual(t, b, body)
	assert.Equal(t, b, gz)
}

func TestDecompressNoContent(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Decompress()(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, GZIPEncoding, req.Header.Get(echo.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func TestDecompressErrorReturned(t *testing.T) {
	e := echo.New()
	r := NewRouter()

	r.GET("/", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return echo.ErrNotFound
		}
	})

	e.Use(Decompress())
	e.Use(r.Routes())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
}

func TestDecompressSkipper(t *testing.T) {
	e := echo.New()
	e.Use(DecompressWithConfig(DecompressConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/skip"
		},
	}))
	body := `{"name": "echo"}`
	req := httptest.NewRequest(http.MethodPost, "/skip", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	reqBody, err := ioutil.ReadAll(c.Request().Body)
	assert.NoError(t, err)
	assert.Equal(t, body, string(reqBody))
}

type TestDecompressPoolWithError struct {
}

func (d *TestDecompressPoolWithError) gzipDecompressPool() sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			return errors.New("pool error")
		},
	}
}

func TestDecompressPoolError(t *testing.T) {
	e := echo.New()
	e.Use(DecompressWithConfig(DecompressConfig{
		Skipper:            DefaultSkipper,
		GzipDecompressPool: &TestDecompressPoolWithError{},
	}))
	body := `{"name": "echo"}`
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	assert.Equal(t, GZIPEncoding, req.Header.Get(echo.HeaderContentEncoding))
	reqBody, err := ioutil.ReadAll(c.Request().Body)
	assert.NoError(t, err)
	assert.Equal(t, body, string(reqBody))
	assert.Equal(t, rec.Code, http.StatusInternalServerError)
}

func BenchmarkDecompress(b *testing.B) {
	e := echo.New()
	body := `{"name": "echo"}`
	gz, _ := gzipString(body)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(gz)))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)

	h := Decompress()(func(c echo.Context) error {
		c.Response().Write([]byte(body)) // For Content-Type sniffing
		return nil
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Decompress
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h(c)
	}
}

func gzipString(body string) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err := gz.Write([]byte(body))
	if err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
