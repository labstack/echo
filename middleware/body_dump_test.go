// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestBodyDump(t *testing.T) {
	e := echo.New()
	hw := "Hello, World!"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c *echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	}

	requestBody := ""
	responseBody := ""
	mw, err := BodyDumpConfig{Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
		requestBody = string(reqBody)
		responseBody = string(resBody)
	}}.ToMiddleware()
	assert.NoError(t, err)

	if assert.NoError(t, mw(h)(c)) {
		assert.Equal(t, requestBody, hw)
		assert.Equal(t, responseBody, hw)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, hw, rec.Body.String())
	}

}

func TestBodyDump_skipper(t *testing.T) {
	e := echo.New()

	isCalled := false
	mw, err := BodyDumpConfig{
		Skipper: func(c *echo.Context) bool {
			return true
		},
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			isCalled = true
		},
	}.ToMiddleware()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c *echo.Context) error {
		return errors.New("some error")
	}

	err = mw(h)(c)
	assert.EqualError(t, err, "some error")
	assert.False(t, isCalled)
}

func TestBodyDump_fails(t *testing.T) {
	e := echo.New()
	hw := "Hello, World!"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c *echo.Context) error {
		return errors.New("some error")
	}

	mw, err := BodyDumpConfig{Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {}}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.EqualError(t, err, "some error")
	assert.Equal(t, http.StatusOK, rec.Code)

}

func TestBodyDumpWithConfig_panic(t *testing.T) {
	assert.Panics(t, func() {
		mw := BodyDumpWithConfig(BodyDumpConfig{
			Skipper: nil,
			Handler: nil,
		})
		assert.NotNil(t, mw)
	})

	assert.NotPanics(t, func() {
		mw := BodyDumpWithConfig(BodyDumpConfig{Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {}})
		assert.NotNil(t, mw)
	})
}

func TestBodyDump_panic(t *testing.T) {
	assert.Panics(t, func() {
		mw := BodyDump(nil)
		assert.NotNil(t, mw)
	})

	assert.NotPanics(t, func() {
		BodyDump(func(c *echo.Context, reqBody, resBody []byte, err error) {})
	})
}

func TestBodyDumpResponseWriter_CanNotFlush(t *testing.T) {
	bdrw := bodyDumpResponseWriter{
		ResponseWriter: new(testResponseWriterNoFlushHijack), // this RW does not support flush
	}
	assert.PanicsWithError(t, "response writer flushing is not supported", func() {
		bdrw.Flush()
	})
}

func TestBodyDumpResponseWriter_CanFlush(t *testing.T) {
	trwu := testResponseWriterUnwrapperHijack{testResponseWriterUnwrapper: testResponseWriterUnwrapper{rw: httptest.NewRecorder()}}
	bdrw := bodyDumpResponseWriter{
		ResponseWriter: &trwu,
	}
	bdrw.Flush()
	assert.Equal(t, 1, trwu.unwrapCalled)
}

func TestBodyDumpResponseWriter_CanUnwrap(t *testing.T) {
	trwu := &testResponseWriterUnwrapper{rw: httptest.NewRecorder()}
	bdrw := bodyDumpResponseWriter{
		ResponseWriter: trwu,
	}
	result := bdrw.Unwrap()
	assert.Equal(t, trwu, result)
}

func TestBodyDumpResponseWriter_CanHijack(t *testing.T) {
	trwu := testResponseWriterUnwrapperHijack{testResponseWriterUnwrapper: testResponseWriterUnwrapper{rw: httptest.NewRecorder()}}
	bdrw := bodyDumpResponseWriter{
		ResponseWriter: &trwu, // this RW supports hijacking through unwrapping
	}
	_, _, err := bdrw.Hijack()
	assert.EqualError(t, err, "can hijack")
}

func TestBodyDumpResponseWriter_CanNotHijack(t *testing.T) {
	trwu := testResponseWriterUnwrapper{rw: httptest.NewRecorder()}
	bdrw := bodyDumpResponseWriter{
		ResponseWriter: &trwu, // this RW supports hijacking through unwrapping
	}
	_, _, err := bdrw.Hijack()
	assert.EqualError(t, err, "feature not supported")
}

func TestBodyDump_ReadError(t *testing.T) {
	e := echo.New()

	// Create a reader that fails during read
	failingReader := &failingReadCloser{
		data:     []byte("partial data"),
		failAt:   7, // Fail after 7 bytes
		failWith: errors.New("connection reset"),
	}

	req := httptest.NewRequest(http.MethodPost, "/", failingReader)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		// This handler should not be reached if body read fails
		body, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(body))
	}

	requestBodyReceived := ""
	mw := BodyDump(func(c *echo.Context, reqBody, resBody []byte, err error) {
		requestBodyReceived = string(reqBody)
	})

	err := mw(h)(c)

	// Verify error is propagated
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection reset")

	// Verify handler was not executed (callback wouldn't have received data)
	assert.Empty(t, requestBodyReceived)
}

// failingReadCloser is a helper type for testing read errors
type failingReadCloser struct {
	data     []byte
	pos      int
	failAt   int
	failWith error
}

func (f *failingReadCloser) Read(p []byte) (n int, err error) {
	if f.pos >= f.failAt {
		return 0, f.failWith
	}

	n = copy(p, f.data[f.pos:])
	f.pos += n

	if f.pos >= f.failAt {
		return n, f.failWith
	}

	return n, nil
}

func (f *failingReadCloser) Close() error {
	return nil
}

func TestBodyDump_RequestWithinLimit(t *testing.T) {
	e := echo.New()
	requestData := "Hello, World!"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(requestData))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		body, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(body))
	}

	requestBodyDumped := ""
	mw, err := BodyDumpConfig{
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			requestBodyDumped = string(reqBody)
		},
		MaxRequestBytes:  1 * MB, // 1MB limit
		MaxResponseBytes: 1 * MB,
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, requestData, requestBodyDumped, "Small request should be fully dumped")
	assert.Equal(t, requestData, rec.Body.String(), "Handler should receive full request")
}

func TestBodyDump_RequestExceedsLimit(t *testing.T) {
	e := echo.New()
	// Create 2KB of data but limit to 1KB
	largeData := strings.Repeat("A", 2*1024)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeData))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		body, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(body))
	}

	requestBodyDumped := ""
	limit := int64(1024) // 1KB limit
	mw, err := BodyDumpConfig{
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			requestBodyDumped = string(reqBody)
		},
		MaxRequestBytes:  limit,
		MaxResponseBytes: 1 * MB,
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, int(limit), len(requestBodyDumped), "Dumped request should be truncated to limit")
	assert.Equal(t, strings.Repeat("A", 1024), requestBodyDumped, "Dumped data should match first N bytes")
	// Handler should receive truncated data (what was dumped)
	assert.Equal(t, strings.Repeat("A", 1024), rec.Body.String())
}

func TestBodyDump_RequestAtExactLimit(t *testing.T) {
	e := echo.New()
	exactData := strings.Repeat("B", 1024)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(exactData))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		body, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(body))
	}

	requestBodyDumped := ""
	limit := int64(1024)
	mw, err := BodyDumpConfig{
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			requestBodyDumped = string(reqBody)
		},
		MaxRequestBytes:  limit,
		MaxResponseBytes: 1 * MB,
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, int(limit), len(requestBodyDumped), "Exact limit should dump full data")
	assert.Equal(t, exactData, requestBodyDumped)
}

func TestBodyDump_ResponseWithinLimit(t *testing.T) {
	e := echo.New()
	responseData := "Response data"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, responseData)
	}

	responseBodyDumped := ""
	mw, err := BodyDumpConfig{
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			responseBodyDumped = string(resBody)
		},
		MaxRequestBytes:  1 * MB,
		MaxResponseBytes: 1 * MB,
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, responseData, responseBodyDumped, "Small response should be fully dumped")
	assert.Equal(t, responseData, rec.Body.String(), "Client should receive full response")
}

func TestBodyDump_ResponseExceedsLimit(t *testing.T) {
	e := echo.New()
	largeResponse := strings.Repeat("X", 2*1024) // 2KB
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, largeResponse)
	}

	responseBodyDumped := ""
	limit := int64(1024) // 1KB limit
	mw, err := BodyDumpConfig{
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			responseBodyDumped = string(resBody)
		},
		MaxRequestBytes:  1 * MB,
		MaxResponseBytes: limit,
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)
	// Dump should be truncated
	assert.Equal(t, int(limit), len(responseBodyDumped), "Dumped response should be truncated to limit")
	assert.Equal(t, strings.Repeat("X", 1024), responseBodyDumped)
	// Client should still receive full response!
	assert.Equal(t, largeResponse, rec.Body.String(), "Client must receive full response despite dump truncation")
}

func TestBodyDump_ClientGetsFullResponse(t *testing.T) {
	e := echo.New()
	// This is critical - even when dump is limited, client gets everything
	largeResponse := strings.Repeat("DATA", 500) // 2KB
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		// Write response in chunks to test incremental writes
		for i := 0; i < 4; i++ {
			c.Response().Write([]byte(strings.Repeat("DATA", 125)))
		}
		return nil
	}

	responseBodyDumped := ""
	mw, err := BodyDumpConfig{
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			responseBodyDumped = string(resBody)
		},
		MaxRequestBytes:  1 * MB,
		MaxResponseBytes: 512, // Very small limit
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, 512, len(responseBodyDumped), "Dump should be limited")
	assert.Equal(t, largeResponse, rec.Body.String(), "Client must get full response")
}

func TestBodyDump_BothLimitsSimultaneous(t *testing.T) {
	e := echo.New()
	largeRequest := strings.Repeat("Q", 2*1024)
	largeResponse := strings.Repeat("R", 2*1024)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeRequest))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		io.ReadAll(c.Request().Body) // Consume request
		return c.String(http.StatusOK, largeResponse)
	}

	requestBodyDumped := ""
	responseBodyDumped := ""
	limit := int64(1024)
	mw, err := BodyDumpConfig{
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			requestBodyDumped = string(reqBody)
			responseBodyDumped = string(resBody)
		},
		MaxRequestBytes:  limit,
		MaxResponseBytes: limit,
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, int(limit), len(requestBodyDumped), "Request dump should be limited")
	assert.Equal(t, int(limit), len(responseBodyDumped), "Response dump should be limited")
}

func TestBodyDump_DefaultConfig(t *testing.T) {
	e := echo.New()
	smallData := "test"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(smallData))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		body, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(body))
	}

	requestBodyDumped := ""
	// Use default config which should have 1MB limits
	config := BodyDumpConfig{}
	config.Handler = func(c *echo.Context, reqBody, resBody []byte, err error) {
		requestBodyDumped = string(reqBody)
	}
	mw, err := config.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)
	assert.Equal(t, smallData, requestBodyDumped)
}

func TestBodyDump_LargeRequestDosPrevention(t *testing.T) {
	e := echo.New()
	// Simulate a very large request (10MB) that could cause OOM
	largeSize := 10 * 1024 * 1024 // 10MB
	largeData := strings.Repeat("M", largeSize)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeData))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		body, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(body))
	}

	requestBodyDumped := ""
	limit := int64(1 * MB) // Only dump 1MB max
	mw, err := BodyDumpConfig{
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			requestBodyDumped = string(reqBody)
		},
		MaxRequestBytes:  limit,
		MaxResponseBytes: 1 * MB,
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)
	// Verify only 1MB was dumped, not 10MB
	assert.Equal(t, int(limit), len(requestBodyDumped), "Should only dump up to limit, preventing DoS")
	assert.Less(t, len(requestBodyDumped), largeSize, "Dump should be much smaller than full request")
}

func TestBodyDump_LargeResponseDosPrevention(t *testing.T) {
	e := echo.New()
	// Simulate a very large response (10MB)
	largeSize := 10 * 1024 * 1024 // 10MB
	largeResponse := strings.Repeat("R", largeSize)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, largeResponse)
	}

	responseBodyDumped := ""
	limit := int64(1 * MB) // Only dump 1MB max
	mw, err := BodyDumpConfig{
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			responseBodyDumped = string(resBody)
		},
		MaxRequestBytes:  1 * MB,
		MaxResponseBytes: limit,
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)
	// Verify only 1MB was dumped, not 10MB
	assert.Equal(t, int(limit), len(responseBodyDumped), "Should only dump up to limit, preventing DoS")
	assert.Less(t, len(responseBodyDumped), largeSize, "Dump should be much smaller than full response")
	// Client still gets full response
	assert.Equal(t, largeSize, rec.Body.Len(), "Client must receive full response")
}

func BenchmarkBodyDump_WithLimit(b *testing.B) {
	e := echo.New()
	requestData := strings.Repeat("data", 256)  // 1KB
	responseData := strings.Repeat("resp", 256) // 1KB

	h := func(c *echo.Context) error {
		io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, responseData)
	}

	mw, _ := BodyDumpConfig{
		Handler: func(c *echo.Context, reqBody, resBody []byte, err error) {
			// Simulate logging
			_ = len(reqBody) + len(resBody)
		},
		MaxRequestBytes:  1 * MB,
		MaxResponseBytes: 1 * MB,
	}.ToMiddleware()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(requestData))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		mw(h)(c)
	}
}

func BenchmarkBodyDump_BufferPooling(b *testing.B) {
	e := echo.New()
	requestData := strings.Repeat("x", 1024)
	responseData := "response"

	h := func(c *echo.Context) error {
		io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, responseData)
	}

	mw, _ := BodyDumpConfig{
		Handler:          func(c *echo.Context, reqBody, resBody []byte, err error) {},
		MaxRequestBytes:  1 * MB,
		MaxResponseBytes: 1 * MB,
	}.ToMiddleware()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(requestData))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		mw(h)(c)
	}
}
