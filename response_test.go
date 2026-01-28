// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	res := NewResponse(rec, e.Logger)

	// Before
	res.Before(func() {
		c.Response().Header().Set(HeaderServer, "echo")
	})
	// After
	res.After(func() {
		c.Response().Header().Set(HeaderXFrameOptions, "DENY")
	})
	res.Write([]byte("test"))
	assert.Equal(t, "echo", rec.Header().Get(HeaderServer))
	assert.Equal(t, "DENY", rec.Header().Get(HeaderXFrameOptions))
}

func TestResponse_Write_FallsBackToDefaultStatus(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	res := NewResponse(rec, e.Logger)

	res.Write([]byte("test"))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestResponse_Write_UsesSetResponseCode(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	res := NewResponse(rec, e.Logger)

	res.Status = http.StatusBadRequest
	res.Write([]byte("test"))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestResponse_ChangeStatusCodeBeforeWrite(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	res := NewResponse(rec, e.Logger)

	res.Before(func() {
		if 200 < res.Status && res.Status < 300 {
			res.Status = 200
		}
	})

	res.WriteHeader(209)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestResponse_Unwrap(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	res := NewResponse(rec, e.Logger)

	assert.Equal(t, rec, res.Unwrap())
}

func TestResponse_isHijacker(t *testing.T) {
	var resp http.ResponseWriter = &Response{}

	_, ok := resp.(http.Hijacker)
	assert.True(t, ok)
}

func TestResponse_Flush(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	res := NewResponse(rec, e.Logger)

	res.Write([]byte("test"))
	res.Flush()
	assert.True(t, rec.Flushed)
}

type testResponseWriter struct {
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
}

func (w *testResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *testResponseWriter) Header() http.Header {
	return nil
}

func TestResponse_FlushPanics(t *testing.T) {
	e := New()
	rw := new(testResponseWriter)
	res := NewResponse(rw, e.Logger)

	// we test that we behave as before unwrapping flushers - flushing writer that does not support it causes panic
	assert.PanicsWithError(t, "echo: response writer *echo.testResponseWriter does not support flushing (http.Flusher interface)", func() {
		res.Flush()
	})
}

func TestResponse_UnwrapResponse(t *testing.T) {
	orgRes := NewResponse(httptest.NewRecorder(), nil)
	res, err := UnwrapResponse(orgRes)

	assert.NotNil(t, res)
	assert.NoError(t, err)
}

func TestResponse_UnwrapResponse_error(t *testing.T) {
	rw := new(testResponseWriter)
	res, err := UnwrapResponse(rw)

	assert.Nil(t, res)
	assert.EqualError(t, err, "ResponseWriter does not implement 'Unwrap() http.ResponseWriter' interface or unwrap to *echo.Response")
}
