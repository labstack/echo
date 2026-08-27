// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestBodyLimitConfig_ToMiddleware(t *testing.T) {
	e := echo.New()
	hw := []byte("Hello, World!")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c *echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	}

	// Based on content length (within limit)
	mw, err := BodyLimitConfig{LimitBytes: 2 * MB}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, hw, rec.Body.Bytes())
	}

	// Based on content read (overlimit)
	mw, err = BodyLimitConfig{LimitBytes: 2}.ToMiddleware()
	assert.NoError(t, err)
	he := mw(h)(c).(echo.HTTPStatusCoder)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())

	// Based on content read (within limit)
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(hw))
	req.ContentLength = -1
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	mw, err = BodyLimitConfig{LimitBytes: 2 * MB}.ToMiddleware()
	assert.NoError(t, err)
	err = mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Hello, World!", rec.Body.String())

	// Based on content read (overlimit)
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(hw))
	req.ContentLength = -1
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	mw, err = BodyLimitConfig{LimitBytes: 2}.ToMiddleware()
	assert.NoError(t, err)
	he = mw(h)(c).(echo.HTTPStatusCoder)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
}

func TestBodyLimitAfterDecompressUsesDecodedSize(t *testing.T) {
	e := echo.New()
	body := "ok"
	gz, err := gzipString(body)
	assert.NoError(t, err)
	assert.Greater(t, len(gz), len(body))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(gz))
	req.Header.Set(echo.HeaderContentEncoding, GZIPEncoding)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = Decompress()(BodyLimit(int64(len(body)))(func(c *echo.Context) error {
		body, readErr := io.ReadAll(c.Request().Body)
		if readErr != nil {
			return readErr
		}
		return c.String(http.StatusOK, string(body))
	}))(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, body, rec.Body.String())
}

func TestBodyLimitOversizedReadDoesNotBypassLimit(t *testing.T) {
	// Callers that process n>0 before treating err as fatal, as io.Reader
	// documents, must still never receive more than LimitBytes.
	tests := []struct {
		name    string
		bufSize int
	}{
		{name: "single oversized read", bufSize: 64},
		{name: "one byte at a time", bufSize: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const limit int64 = 5
			body := bytes.Repeat([]byte("x"), 10*int(limit))

			e := echo.New()
			var total int
			var lastErr error
			e.POST("/", func(c *echo.Context) error {
				buf := make([]byte, tc.bufSize)
				for {
					n, err := c.Request().Body.Read(buf)
					total += n
					lastErr = err
					if n == 0 {
						break
					}
				}
				return c.NoContent(http.StatusOK)
			}, BodyLimit(limit))

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			req.ContentLength = -1
			req.TransferEncoding = []string{"chunked"}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, int(limit), total)
			he, ok := lastErr.(echo.HTTPStatusCoder)
			if assert.True(t, ok) {
				assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
			}
		})
	}
}

type countingReadCloser struct {
	io.Reader
	read  int64
	calls int
}

func (c *countingReadCloser) Read(b []byte) (int, error) {
	c.calls++
	n, err := c.Reader.Read(b)
	c.read += int64(n)
	return n, err
}

func (c *countingReadCloser) Close() error { return nil }

func TestBodyLimitDoesNotOverdrawTheSource(t *testing.T) {
	const limit int64 = 5
	src := &countingReadCloser{Reader: bytes.NewReader(bytes.Repeat([]byte("x"), 1<<20))}

	e := echo.New()
	var callsAtRefusal int
	e.POST("/", func(c *echo.Context) error {
		buf := make([]byte, 64*1024)
		_, _ = c.Request().Body.Read(buf)
		callsAtRefusal = src.calls
		for i := 0; i < 5; i++ {
			_, _ = c.Request().Body.Read(buf)
		}
		return c.NoContent(http.StatusOK)
	}, BodyLimit(limit))

	req := httptest.NewRequest(http.MethodPost, "/", src)
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.LessOrEqual(t, src.read, limit+1)
	assert.Equal(t, callsAtRefusal, src.calls)
}

func TestBodyLimitReadStaysRefused(t *testing.T) {
	e := echo.New()
	e.POST("/", func(c *echo.Context) error {
		_, err := io.ReadAll(c.Request().Body)
		he, ok := err.(echo.HTTPStatusCoder)
		if assert.True(t, ok) {
			assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
		}

		n, err := c.Request().Body.Read(make([]byte, 8))
		assert.Equal(t, 0, n)
		he, ok = err.(echo.HTTPStatusCoder)
		if assert.True(t, ok) {
			assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
		}
		return c.NoContent(http.StatusOK)
	}, BodyLimit(2))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("Hello, World!")))
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestBodyLimitJSONDecoderDoesNotAcceptOversizedBody(t *testing.T) {
	const limit int64 = 16
	payload := `{"data":"` + strings.Repeat("x", 1024) + `"}`

	e := echo.New()
	var decodeErr error
	var got string
	e.POST("/", func(c *echo.Context) error {
		var v struct {
			Data string `json:"data"`
		}
		decodeErr = json.NewDecoder(c.Request().Body).Decode(&v)
		got = v.Data
		if decodeErr != nil {
			return decodeErr
		}
		return c.NoContent(http.StatusOK)
	}, BodyLimit(limit))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.NotEqual(t, strings.Repeat("x", 1024), got)
	var he echo.HTTPStatusCoder
	if assert.True(t, errors.As(decodeErr, &he), "decode error: %v", decodeErr) {
		assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
	}
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}

func TestBodyLimitResetAfterOversizeRequest(t *testing.T) {
	e := echo.New()
	e.POST("/", func(c *echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(b))
	}, BodyLimit(5))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bytes.Repeat([]byte("x"), 20)))
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("hello")))
	req.ContentLength = -1
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "hello", rec.Body.String())
}

func TestBodyLimitReader(t *testing.T) {
	hw := []byte("Hello, World!")

	config := BodyLimitConfig{
		Skipper:    DefaultSkipper,
		LimitBytes: 2,
	}
	reader := &limitedReader{
		BodyLimitConfig: config,
		reader:          io.NopCloser(bytes.NewReader(hw)),
	}

	// read all should return ErrStatusRequestEntityTooLarge
	_, err := io.ReadAll(reader)
	he := err.(echo.HTTPStatusCoder)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())

	// reset reader and read two bytes must succeed
	bt := make([]byte, 2)
	reader.Reset(io.NopCloser(bytes.NewReader(hw)))
	n, err := reader.Read(bt)
	assert.Equal(t, 2, n)
	assert.Equal(t, nil, err)
}

func TestBodyLimit_skipper(t *testing.T) {
	e := echo.New()
	h := func(c *echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	}
	mw, err := BodyLimitConfig{
		Skipper: func(c *echo.Context) bool {
			return true
		},
		LimitBytes: 2,
	}.ToMiddleware()
	assert.NoError(t, err)

	hw := []byte("Hello, World!")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, hw, rec.Body.Bytes())
}

func TestBodyLimitWithConfig(t *testing.T) {
	e := echo.New()
	hw := []byte("Hello, World!")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c *echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	}

	mw := BodyLimitWithConfig(BodyLimitConfig{LimitBytes: 2 * MB})

	err := mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, hw, rec.Body.Bytes())
}

func TestBodyLimit(t *testing.T) {
	e := echo.New()
	hw := []byte("Hello, World!")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c *echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	}

	mw := BodyLimit(2 * MB)

	err := mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, hw, rec.Body.Bytes())
}
