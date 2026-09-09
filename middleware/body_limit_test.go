// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestBodyLimitReader_stickyError(t *testing.T) {
	// Verifies that once the limit is exceeded, subsequent Read calls always
	// return (0, err) — no more data is leaked to the caller.
	// This is the fix for issue #3071 where callers following the standard
	// io.Reader pattern (process n>0 before checking err) could read arbitrarily
	// far past the limit.
	hw := []byte("Hello, World!")

	config := BodyLimitConfig{
		Skipper:    DefaultSkipper,
		LimitBytes: 5,
	}
	reader := &limitedReader{
		BodyLimitConfig: config,
		reader:          io.NopCloser(bytes.NewReader(hw)),
	}

	buf := make([]byte, 1)

	// Read 5 bytes one at a time — should all succeed with no error
	for i := 0; i < 5; i++ {
		n, err := reader.Read(buf)
		assert.Equal(t, 1, n)
		assert.NoError(t, err)
	}

	// 6th read should cross the limit and return the error alongside 1 byte
	n, err := reader.Read(buf)
	assert.Equal(t, 1, n)
	he := err.(echo.HTTPStatusCoder)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())

	// All subsequent reads must return (0, err) — never more data
	for i := 0; i < 10; i++ {
		n, err = reader.Read(buf)
		assert.Equal(t, 0, n, "read after limit exceeded must return 0 bytes")
		he = err.(echo.HTTPStatusCoder)
		assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
	}
}

func TestBodyLimitReader_boundedRead(t *testing.T) {
	// Verifies that a single large Read buffer can never return enough bytes to
	// both cross the limit and complete a value — only remaining+1 bytes max.
	// This prevents a single oversized Read from silently bypassing the limit.
	hw := bytes.Repeat([]byte("x"), 100)

	config := BodyLimitConfig{
		Skipper:    DefaultSkipper,
		LimitBytes: 10,
	}
	reader := &limitedReader{
		BodyLimitConfig: config,
		reader:          io.NopCloser(bytes.NewReader(hw)),
	}

	// A large buffer read should return at most remaining+1 = 11 bytes, not the full 100
	buf := make([]byte, 100)
	n, err := reader.Read(buf)
	assert.LessOrEqual(t, n, 11, "single read must not return more than remaining+1 bytes")
	he := err.(echo.HTTPStatusCoder)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
}

func TestBodyLimit_readAfterLimitNoLeak(t *testing.T) {
	// Integration test: verifies that BodyLimit middleware correctly enforces the
	// limit even when the handler reads with a large buffer and processes n>0
	// bytes before checking err (standard io.Reader semantics).
	e := echo.New()

	const limit = 5
	body := bytes.Repeat([]byte("x"), 10*limit) // 50 bytes, well over the limit

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	// Force chunked encoding to bypass the Content-Length fast path
	req.ContentLength = -1
	req.TransferEncoding = []string{"chunked"}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := BodyLimit(limit)
	h := func(c *echo.Context) error {
		// Simulate a handler that follows io.Reader semantics exactly:
		// process n>0 bytes before considering err fatal.
		// (This is exactly what encoding/json.Decoder does.)
		buf := make([]byte, 64)
		var total int
		for {
			n, err := c.Request().Body.Read(buf)
			total += n
			if n == 0 {
				break
			}
			_ = err // process data first, per io.Reader docs
		}
		// The handler must never be able to read more than limit+1 bytes total.
		assert.LessOrEqual(t, total, limit+1, "handler must not read past limit+1")
		return nil
	}

	err := mw(h)(c)
	he := err.(echo.HTTPStatusCoder)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
}
