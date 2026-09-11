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

func TestBodyLimitReader_singleOversizedRead(t *testing.T) {
	hw := []byte("Hello, World!")
	reader := &limitedReader{
		BodyLimitConfig: BodyLimitConfig{LimitBytes: 2},
		reader:          io.NopCloser(bytes.NewReader(hw)),
	}

	// a single read with a buffer much larger than the limit must deliver at
	// most the allowed bytes and report the limit as exceeded
	buf := make([]byte, 64)
	n, err := reader.Read(buf)
	assert.Equal(t, 2, n)
	he := err.(echo.HTTPStatusCoder)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
}

func TestBodyLimitReader_noDataAfterLimitExceeded(t *testing.T) {
	hw := bytes.Repeat([]byte("x"), 64)
	reader := &limitedReader{
		BodyLimitConfig: BodyLimitConfig{LimitBytes: 5},
		reader:          io.NopCloser(bytes.NewReader(hw)),
	}

	// a caller following the io.Reader contract processes the n>0 bytes
	// before considering the error and keeps calling Read; it must never
	// receive more data once the limit has been exceeded
	buf := make([]byte, 64)
	total := 0
	var err error
	for {
		var n int
		n, err = reader.Read(buf)
		total += n
		if n == 0 {
			break
		}
	}

	assert.Equal(t, 5, total)
	he := err.(echo.HTTPStatusCoder)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.StatusCode())
}

func TestBodyLimitReader_exactLimitBody(t *testing.T) {
	hw := []byte("ab")
	reader := &limitedReader{
		BodyLimitConfig: BodyLimitConfig{LimitBytes: 2},
		reader:          io.NopCloser(bytes.NewReader(hw)),
	}

	// a body of exactly the limit size is not over the limit and must be
	// readable completely
	data, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "ab", string(data))
}

func TestBodyLimit_oversizedBodyWithContractCompliantReader(t *testing.T) {
	e := echo.New()
	const limit = 5
	h := func(c *echo.Context) error {
		buf := make([]byte, 64)
		total := 0
		for {
			n, err := c.Request().Body.Read(buf)
			total += n
			if n == 0 {
				break
			}
			// process the n>0 bytes before considering the error, exactly
			// what io.Reader's documentation tells callers to do
			_ = err
		}
		assert.LessOrEqual(t, total, limit)
		return c.String(http.StatusOK, "ok")
	}
	mw, err := BodyLimitConfig{LimitBytes: limit}.ToMiddleware()
	assert.NoError(t, err)

	body := bytes.Repeat([]byte("x"), 10*limit)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = -1 // force the content-read path
	req.TransferEncoding = []string{"chunked"}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
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
