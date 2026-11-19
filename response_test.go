// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"fmt"
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

func TestResponse_FlushBeforeWrite(t *testing.T) {
	for _, status := range []int{0, http.StatusAccepted} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			rec := httptest.NewRecorder()
			res := NewResponse(rec, New().Logger)
			res.Status = status
			beforeCalls := 0
			res.Before(func() {
				beforeCalls++
				res.Header().Set("X-Before", "called")
			})

			res.Flush()
			wantStatus := status
			if wantStatus == 0 {
				wantStatus = http.StatusOK
			}
			assert.Equal(t, wantStatus, rec.Code)
			assert.Equal(t, wantStatus, res.Status)
			assert.True(t, res.Committed)
			assert.True(t, rec.Flushed)
			assert.Equal(t, "called", rec.Result().Header.Get("X-Before"))

			res.Flush()
			_, err := res.Write([]byte("streamed"))
			assert.NoError(t, err)
			assert.Equal(t, 1, beforeCalls)
			assert.Equal(t, "streamed", rec.Body.String())
		})
	}
}

func TestResponse_FlushDuringBefore(t *testing.T) {
	for _, trigger := range []string{"Flush", "Write", "WriteHeader"} {
		t.Run(trigger, func(t *testing.T) {
			rec := &flushCountingResponseWriter{ResponseRecorder: httptest.NewRecorder()}
			res := NewResponse(rec, New().Logger)
			res.Status = http.StatusAccepted
			var beforeCalls [2]int
			afterCalls := 0
			res.Before(func() {
				beforeCalls[0]++
				if beforeCalls[0] > 1 {
					t.Fatal("Before callback reentered")
				}
				res.Flush()
				res.Flush()
				assert.False(t, res.Committed)
				assert.False(t, rec.Flushed)
				res.Status = http.StatusCreated
				res.Header().Set("X-First", "after flush")
			})
			res.Before(func() {
				beforeCalls[1]++
				res.Flush()
				assert.False(t, rec.Flushed)
				res.Status = http.StatusNonAuthoritativeInfo
				res.Header().Set("X-Second", "after flush")
			})
			res.After(func() { afterCalls++ })

			switch trigger {
			case "Flush":
				res.Flush()
			case "Write":
				_, err := res.Write([]byte("body"))
				assert.NoError(t, err)
			case "WriteHeader":
				res.WriteHeader(http.StatusAccepted)
			}

			assert.Equal(t, [2]int{1, 1}, beforeCalls)
			assert.True(t, res.Committed)
			assert.Equal(t, http.StatusNonAuthoritativeInfo, res.Status)
			assert.Equal(t, res.Status, rec.Code)
			assert.Equal(t, 1, rec.flushes)
			assert.Equal(t, "after flush", rec.Result().Header.Get("X-First"))
			assert.Equal(t, "after flush", rec.Result().Header.Get("X-Second"))
			if trigger == "Write" {
				assert.Equal(t, "body", rec.Body.String())
				assert.Equal(t, 1, afterCalls)
			} else {
				assert.Equal(t, 0, afterCalls)
			}

			res.Flush()
			assert.Equal(t, 2, rec.flushes)
			assert.Equal(t, [2]int{1, 1}, beforeCalls)
		})
	}
}

func TestResponse_FlushDuringBeforePanic(t *testing.T) {
	rec := &flushCountingResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	res := NewResponse(rec, New().Logger)
	firstCall := true
	res.Before(func() {
		if firstCall {
			firstCall = false
			res.Flush()
			panic("before failed")
		}
	})

	assert.PanicsWithValue(t, "before failed", func() {
		res.WriteHeader(http.StatusAccepted)
	})
	assert.False(t, res.Committed)
	assert.False(t, rec.Flushed)

	res.WriteHeader(http.StatusInternalServerError)
	assert.True(t, res.Committed)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, 0, rec.flushes)
	res.Flush()
	assert.Equal(t, 1, rec.flushes)

	next := &flushCountingResponseWriter{ResponseRecorder: httptest.NewRecorder()}
	res.reset(next)
	res.Flush()
	assert.True(t, res.Committed)
	assert.Equal(t, http.StatusOK, next.Code)
	assert.Equal(t, 1, next.flushes)
}

type flushCountingResponseWriter struct {
	*httptest.ResponseRecorder
	flushes int
}

func (w *flushCountingResponseWriter) Flush() {
	w.flushes++
	w.ResponseRecorder.Flush()
}

func TestDelayedStatusWriter_FlushBeforeWrite(t *testing.T) {
	for _, status := range []int{0, http.StatusCreated} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			rec := httptest.NewRecorder()
			w := &delayedStatusWriter{ResponseWriter: rec, status: status}
			w.Flush()

			wantStatus := status
			if wantStatus == 0 {
				wantStatus = http.StatusOK
			}
			assert.Equal(t, wantStatus, rec.Code)
			assert.True(t, w.committed)
			assert.True(t, rec.Flushed)

			_, err := w.Write([]byte("streamed"))
			assert.NoError(t, err)
			assert.Equal(t, wantStatus, rec.Code)
		})
	}
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

// headResponseWriter unit tests

func TestHeadResponseWriter_Write_CountsBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	n, err := w.Write([]byte("hello"))

	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, int64(5), w.bodyBytes)
	assert.Equal(t, "", rec.Body.String())
}

func TestHeadResponseWriter_Write_MultipleCallsSumBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	w.Write([]byte("foo"))   // 3 bytes
	w.Write([]byte("bar"))   // 3 bytes
	w.Write([]byte("baz!!")) // 5 bytes

	assert.Equal(t, int64(11), w.bodyBytes)
	assert.Equal(t, "", rec.Body.String())
}

func TestHeadResponseWriter_Write_ImpliesStatus200(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	w.Write([]byte("x"))

	assert.True(t, w.wroteStatus)
	assert.Equal(t, http.StatusOK, w.status)
}

func TestHeadResponseWriter_WriteHeader_FirstCallWins(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	w.WriteHeader(http.StatusAccepted)   // 202 — should stick
	w.WriteHeader(http.StatusBadRequest) // 400 — should be ignored

	assert.Equal(t, http.StatusAccepted, w.status)
}

func TestHeadResponseWriter_Commit_SetsContentLength(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello world")) // 11 bytes

	w.commit()

	assert.Equal(t, "11", rec.Header().Get(HeaderContentLength))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHeadResponseWriter_Commit_PreservesExplicitContentLength(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set(HeaderContentLength, "999") // handler set this explicitly
	w := &headResponseWriter{rw: rec}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hi")) // 2 bytes — must NOT overwrite

	w.commit()

	assert.Equal(t, "999", rec.Header().Get(HeaderContentLength))
}

func TestHeadResponseWriter_Commit_NoContentLength_TransferEncoding(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Transfer-Encoding", "chunked")
	w := &headResponseWriter{rw: rec}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("some data"))

	w.commit()

	assert.Equal(t, "", rec.Header().Get(HeaderContentLength))
}

func TestHeadResponseWriter_Commit_NoContentLength_204(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	w.WriteHeader(http.StatusNoContent)
	w.commit()

	assert.Equal(t, "", rec.Header().Get(HeaderContentLength))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHeadResponseWriter_Commit_NoContentLength_304(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	w.WriteHeader(http.StatusNotModified)
	w.commit()

	assert.Equal(t, "", rec.Header().Get(HeaderContentLength))
}

func TestHeadResponseWriter_Commit_NoContentLength_1xx(t *testing.T) {
	for _, code := range []int{100, 101} {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			rec := httptest.NewRecorder()
			w := &headResponseWriter{rw: rec}

			w.WriteHeader(code)
			w.commit()

			assert.Equal(t, "", rec.Header().Get(HeaderContentLength))
		})
	}
}

// TestHeadResponseWriter_Commit_NoWriteHeader_WhenNotCommitted verifies that
// commit() does not set Content-Length or call WriteHeader when the handler
// never wrote a status (error-handler path).
func TestHeadResponseWriter_Commit_NoWriteHeader_WhenNotCommitted(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	w.commit()

	assert.False(t, w.wroteStatus)
	assert.Equal(t, "", rec.Header().Get(HeaderContentLength))
}

func TestHeadResponseWriter_Flush_IsNoop(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	w.WriteHeader(http.StatusOK)
	assert.NotPanics(t, func() { w.Flush() })

	// Headers must not have been committed to the recorder yet
	assert.Equal(t, "", rec.Header().Get(HeaderContentLength))
}

func TestHeadResponseWriter_Header_DelegatesToUnderlying(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	w.Header().Set("X-Test", "value")

	assert.Equal(t, "value", rec.Header().Get("X-Test"))
}

func TestHeadResponseWriter_Unwrap_ReturnsUnderlying(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	assert.Equal(t, rec, w.Unwrap())
}

func TestHeadResponseWriter_Hijack_DelegatesToUnderlying(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &headResponseWriter{rw: rec}

	// httptest.Recorder does not support Hijack; expect an error, not a panic
	_, _, err := w.Hijack()
	assert.Error(t, err)
}

// wrapHeadHandler unit tests

func TestWrapHeadHandler_SuppressesBodyWrittenByHandler(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/", nil)
	c := e.NewContext(req, rec)

	handler := func(c *Context) error {
		return c.String(http.StatusOK, "hello") // 5 bytes
	}
	wrapped := wrapHeadHandler(handler)

	err := wrapped(c)

	assert.NoError(t, err)
	assert.Equal(t, "", rec.Body.String())
	assert.Equal(t, "5", rec.Header().Get(HeaderContentLength))
}

func TestWrapHeadHandler_PreservesCustomResponseHeaders(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/", nil)
	c := e.NewContext(req, rec)

	handler := func(c *Context) error {
		c.Response().Header().Set("X-Custom", "present")
		return c.String(http.StatusOK, "body")
	}
	wrapped := wrapHeadHandler(handler)

	wrapped(c)

	assert.Equal(t, "present", rec.Header().Get("X-Custom"))
}

func TestWrapHeadHandler_RestoresOriginalWriterAfterHandler(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/", nil)
	c := e.NewContext(req, rec)
	original := c.Response()

	handler := func(c *Context) error { return nil }
	wrapHeadHandler(handler)(c)

	assert.Equal(t, original, c.Response())
}

func TestWrapHeadHandler_PropagatesHandlerError(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/", nil)
	c := e.NewContext(req, rec)

	handler := func(c *Context) error {
		return ErrNotFound
	}
	wrapped := wrapHeadHandler(handler)

	err := wrapped(c)

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestWrapHeadHandler_HandlerPanic_OriginalWriterRestored(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/", nil)
	c := e.NewContext(req, rec)
	original := c.Response()

	handler := func(c *Context) error {
		panic("intentional")
	}
	wrapped := wrapHeadHandler(handler)

	assert.Panics(t, func() { wrapped(c) })
	assert.Equal(t, original, c.Response()) // defer must have run
}

func TestHeadResponseWriter_WriteHeader_SetsCommittedOnUnderlying(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	underlying := NewResponse(rec, e.Logger)
	w := &headResponseWriter{rw: underlying}

	w.WriteHeader(http.StatusOK)

	assert.True(t, underlying.Committed)
}

func TestWrapHeadHandler_RouteLevelMiddlewareSeesTrueCommitted(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/", nil)
	c := e.NewContext(req, rec)

	var committedAfterHandler bool

	mw := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			err := next(c)
			if r, unwrapErr := UnwrapResponse(c.Response()); unwrapErr == nil {
				committedAfterHandler = r.Committed
			}
			return err
		}
	}

	chain := mw(func(c *Context) error {
		return c.String(http.StatusOK, "hello")
	})
	wrapped := wrapHeadHandler(chain)

	assert.NoError(t, wrapped(c))
	assert.True(t, committedAfterHandler)
}
