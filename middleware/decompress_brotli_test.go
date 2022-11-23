package middleware

import (
	"bytes"
	"github.com/andybalholm/brotli"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDecompressBrotli(t *testing.T) {
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

	assert.Equal(t, "test", rec.Body.String())

	// Decompress
	body := `{"name": "echo"}`
	gz, _ := brotliString(body)
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(gz)))
	req.Header.Set(echo.HeaderContentEncoding, BrotliEncoding)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h(c)
	assert.Equal(t, BrotliEncoding, req.Header.Get(echo.HeaderContentEncoding))
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.Equal(t, body, string(b))
}

func TestDecompressBrotliDefaultConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := DecompressWithConfig(DecompressConfig{})(func(c echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})
	h(c)

	assert.Equal(t, "test", rec.Body.String())

	// Decompress
	body := `{"name": "echo"}`
	gz, _ := brotliString(body)
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(gz)))
	req.Header.Set(echo.HeaderContentEncoding, BrotliEncoding)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h(c)
	assert.Equal(t, BrotliEncoding, req.Header.Get(echo.HeaderContentEncoding))
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.Equal(t, body, string(b))
}

func TestCompressBrotliRequestWithoutDecompressMiddleware(t *testing.T) {
	e := echo.New()
	body := `{"name":"echo"}`
	gz, _ := brotliString(body)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(gz)))
	req.Header.Set(echo.HeaderContentEncoding, BrotliEncoding)
	rec := httptest.NewRecorder()
	e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	assert.Equal(t, BrotliEncoding, req.Header.Get(echo.HeaderContentEncoding))
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.NotEqual(t, b, body)
	assert.Equal(t, b, gz)
}

func TestDecompressBrotliNoContent(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentEncoding, BrotliEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Decompress()(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, BrotliEncoding, req.Header.Get(echo.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func TestDecompressBrotliErrorReturned(t *testing.T) {
	e := echo.New()
	e.Use(Decompress())
	e.GET("/", func(c echo.Context) error {
		return echo.ErrNotFound
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentEncoding, BrotliEncoding)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
}

func TestDecompressBrotliSkipper(t *testing.T) {
	e := echo.New()
	e.Use(DecompressWithConfig(DecompressConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/skip"
		},
	}))
	body := `{"name": "echo"}`
	req := httptest.NewRequest(http.MethodPost, "/skip", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentEncoding, BrotliEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.ServeHTTP(rec, req)
	assert.Equal(t, rec.Header().Get(echo.HeaderContentType), echo.MIMEApplicationJSONCharsetUTF8)
	reqBody, err := ioutil.ReadAll(c.Request().Body)
	assert.NoError(t, err)
	assert.Equal(t, body, string(reqBody))
}

func BenchmarkDecompressBrotli(b *testing.B) {
	e := echo.New()
	body := `{"name": "echo"}`
	bz, _ := brotliString(body)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(bz)))
	req.Header.Set(echo.HeaderContentEncoding, BrotliEncoding)

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

func brotliString(body string) ([]byte, error) {
	var buf bytes.Buffer
	bz := brotli.NewWriter(&buf)

	_, err := bz.Write([]byte(body))
	if err != nil {
		return nil, err
	}

	if err := bz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
