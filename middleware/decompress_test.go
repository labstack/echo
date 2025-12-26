// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestDecompress(t *testing.T) {
	e := echo.New()

	h := Decompress()(func(c *echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})

	// Decompress request body
	body := `{"name": "echo"}`
	gz, _ := gzipString(body)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(gz)))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	assert.NoError(t, err)

	assert.Equal(t, GZIPEncoding, req.Header.Get(echo.HeaderContentEncoding))
	b, err := io.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.Equal(t, body, string(b))
}

func TestDecompress_skippedIfNoHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Skip if no Content-Encoding header
	h := Decompress()(func(c *echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, "test", rec.Body.String())

}

func TestDecompressWithConfig_DefaultConfig_noDecode(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h, err := DecompressConfig{}.ToMiddleware()
	assert.NoError(t, err)

	err = h(func(c *echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})(c)
	assert.NoError(t, err)

	assert.Equal(t, "test", rec.Body.String())

}

func TestDecompressWithConfig_DefaultConfig(t *testing.T) {
	e := echo.New()

	h := Decompress()(func(c *echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})

	// Decompress
	body := `{"name": "echo"}`
	gz, _ := gzipString(body)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(gz)))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	assert.NoError(t, err)

	assert.Equal(t, GZIPEncoding, req.Header.Get(echo.HeaderContentEncoding))
	b, err := io.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.Equal(t, body, string(b))
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
	b, err := io.ReadAll(req.Body)
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
	h := Decompress()(func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	err := h(c)

	if assert.NoError(t, err) {
		assert.Equal(t, GZIPEncoding, req.Header.Get(echo.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func TestDecompressErrorReturned(t *testing.T) {
	e := echo.New()
	e.Use(Decompress())
	e.GET("/", func(c *echo.Context) error {
		return echo.ErrNotFound
	})
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
		Skipper: func(c *echo.Context) bool {
			return c.Request().URL.Path == "/skip"
		},
	}))
	body := `{"name": "echo"}`
	req := httptest.NewRequest(http.MethodPost, "/skip", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	e.ServeHTTP(rec, req)

	assert.Equal(t, rec.Header().Get(echo.HeaderContentType), echo.MIMEApplicationJSON)
	reqBody, err := io.ReadAll(c.Request().Body)
	assert.NoError(t, err)
	assert.Equal(t, body, string(reqBody))
}

type TestDecompressPoolWithError struct {
}

func (d *TestDecompressPoolWithError) gzipDecompressPool() sync.Pool {
	return sync.Pool{
		New: func() any {
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
	reqBody, err := io.ReadAll(c.Request().Body)
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

	h := Decompress()(func(c *echo.Context) error {
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

func TestDecompress_WithinLimit(t *testing.T) {
	e := echo.New()
	body := strings.Repeat("test data ", 100) // Small payload ~1KB
	gz, _ := gzipString(body)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(gz))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h, err := DecompressConfig{MaxDecompressedSize: 100 * MB}.ToMiddleware()
	assert.NoError(t, err)

	err = h(func(c *echo.Context) error {
		b, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(b))
	})(c)

	assert.NoError(t, err)
	assert.Equal(t, body, rec.Body.String())
}

func TestDecompress_ExceedsLimit(t *testing.T) {
	e := echo.New()
	// Create 2KB of data but limit to 1KB
	largeBody := strings.Repeat("A", 2*1024)
	gz, _ := gzipString(largeBody)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(gz))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h, err := DecompressConfig{MaxDecompressedSize: 1024}.ToMiddleware() // 1KB limit
	assert.NoError(t, err)

	err = h(func(c *echo.Context) error {
		_, readErr := io.ReadAll(c.Request().Body)
		return readErr
	})(c)

	// Should return 413 error
	assert.Error(t, err)
	he, ok := err.(echo.HTTPStatusCoder)
	assert.True(t, ok)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
}

func TestDecompress_AtExactLimit(t *testing.T) {
	e := echo.New()
	exactBody := strings.Repeat("B", 1024) // Exactly 1KB
	gz, _ := gzipString(exactBody)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(gz))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h, err := DecompressConfig{MaxDecompressedSize: 1024}.ToMiddleware()
	assert.NoError(t, err)

	err = h(func(c *echo.Context) error {
		b, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(b))
	})(c)

	assert.NoError(t, err)
	assert.Equal(t, exactBody, rec.Body.String())
}

func TestDecompress_ZipBomb(t *testing.T) {
	e := echo.New()
	// Create highly compressed data that expands to 2MB
	// but limit is 1MB
	largeBody := bytes.Repeat([]byte("A"), 2*1024*1024) // 2MB
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	gzWriter.Write(largeBody)
	gzWriter.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h, err := DecompressConfig{MaxDecompressedSize: 1 * MB}.ToMiddleware()
	assert.NoError(t, err)

	err = h(func(c *echo.Context) error {
		_, readErr := io.ReadAll(c.Request().Body)
		return readErr
	})(c)

	// Should return 413 error
	assert.Error(t, err)
	he, ok := err.(echo.HTTPStatusCoder)
	assert.True(t, ok)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
}

func TestDecompress_UnlimitedExplicit(t *testing.T) {
	e := echo.New()
	largeBody := strings.Repeat("X", 10*1024) // 10KB
	gz, _ := gzipString(largeBody)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(gz))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h, err := DecompressConfig{MaxDecompressedSize: -1}.ToMiddleware() // Unlimited
	assert.NoError(t, err)

	err = h(func(c *echo.Context) error {
		b, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(b))
	})(c)

	assert.NoError(t, err)
	assert.Equal(t, largeBody, rec.Body.String())
}

func TestDecompress_DefaultLimit(t *testing.T) {
	e := echo.New()
	smallBody := "test"
	gz, _ := gzipString(smallBody)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(gz))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Use zero value which should apply 100MB default
	h, err := DecompressConfig{}.ToMiddleware()
	assert.NoError(t, err)

	err = h(func(c *echo.Context) error {
		b, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(b))
	})(c)

	assert.NoError(t, err)
	assert.Equal(t, smallBody, rec.Body.String())
}

func TestDecompress_SmallCustomLimit(t *testing.T) {
	e := echo.New()
	body := strings.Repeat("D", 512) // 512 bytes
	gz, _ := gzipString(body)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(gz))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h, err := DecompressConfig{MaxDecompressedSize: 1024}.ToMiddleware() // 1KB limit
	assert.NoError(t, err)

	err = h(func(c *echo.Context) error {
		b, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(b))
	})(c)

	assert.NoError(t, err)
	assert.Equal(t, body, rec.Body.String())
}

func TestDecompress_MultipleReads(t *testing.T) {
	e := echo.New()
	// Test that limit is enforced across multiple Read() calls
	largeBody := strings.Repeat("M", 2*1024) // 2KB
	gz, _ := gzipString(largeBody)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(gz))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h, err := DecompressConfig{MaxDecompressedSize: 1024}.ToMiddleware() // 1KB limit
	assert.NoError(t, err)

	err = h(func(c *echo.Context) error {
		// Read in small chunks
		buf := make([]byte, 256)
		total := 0
		for {
			n, readErr := c.Request().Body.Read(buf)
			total += n
			if readErr != nil {
				if readErr == io.EOF {
					return nil
				}
				return readErr
			}
		}
	})(c)

	// Should return 413 error from cumulative reads
	assert.Error(t, err)
	he, ok := err.(echo.HTTPStatusCoder)
	assert.True(t, ok)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
}

func TestDecompress_LargePayloadDosPrevention(t *testing.T) {
	e := echo.New()
	// Simulate a DoS attack with highly compressed large payload
	largeSize := 10 * 1024 * 1024 // 10MB decompressed
	largeBody := bytes.Repeat([]byte("Z"), largeSize)
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	gzWriter.Write(largeBody)
	gzWriter.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h, err := DecompressConfig{MaxDecompressedSize: 1 * MB}.ToMiddleware()
	assert.NoError(t, err)

	err = h(func(c *echo.Context) error {
		_, readErr := io.ReadAll(c.Request().Body)
		return readErr
	})(c)

	// Should prevent DoS by returning 413
	assert.Error(t, err)
	he, ok := err.(echo.HTTPStatusCoder)
	assert.True(t, ok)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
}

func BenchmarkDecompress_WithLimit(b *testing.B) {
	e := echo.New()
	body := strings.Repeat("benchmark data ", 1000) // ~15KB
	gz, _ := gzipString(body)

	h, _ := DecompressConfig{MaxDecompressedSize: 100 * MB}.ToMiddleware()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(gz))
		req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h(func(c *echo.Context) error {
			io.ReadAll(c.Request().Body)
			return nil
		})(c)
	}
}
