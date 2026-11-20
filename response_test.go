// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestResponse_Flush_CommitsResponse(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	res := NewResponse(rec, e.Logger)
	beforeCalled := false
	res.Before(func() {
		beforeCalled = true
		res.Header().Set("X-Before", "1")
	})

	res.Flush()

	assert.True(t, res.Committed)
	assert.True(t, beforeCalled)
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "1", rec.Header().Get("X-Before"))
	assert.True(t, rec.Flushed)
}

func TestResponse_Flush_ErrorHandlerSkipsCommittedResponse(t *testing.T) {
	e := New()
	e.GET("/", func(c *Context) error {
		if err := http.NewResponseController(c.Response()).Flush(); err != nil {
			return err
		}
		return ErrBadRequest
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestResponse_WriteHeader_Informational(t *testing.T) {
	e := New()
	srv := httptest.NewServer(e)
	defer srv.Close()

	var status int
	var committedAfter1xx bool
	e.GET("/", func(c *Context) error {
		c.Response().Header().Set("Link", "</style.css>; rel=preload; as=style")
		c.Response().WriteHeader(http.StatusEarlyHints)
		r, err := UnwrapResponse(c.Response())
		if err != nil {
			return err
		}
		committedAfter1xx = r.Committed
		err = c.String(http.StatusCreated, "ok")
		status = r.Status
		return err
	})

	var informational []int
	trace := &httptrace.ClientTrace{
		Got1xxResponse: func(code int, _ textproto.MIMEHeader) error {
			informational = append(informational, code)
			return nil
		},
	}
	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(context.Background(), trace), http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, []int{http.StatusEarlyHints}, informational)
	assert.False(t, committedAfter1xx)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "ok", string(body))
	assert.Equal(t, http.StatusCreated, status)
}

func TestResponse_Flush_PanicsBeforeCommit(t *testing.T) {
	e := New()
	res := NewResponse(new(testResponseWriter), e.Logger)
	beforeCalled := false
	res.Before(func() { beforeCalled = true })

	assert.Panics(t, func() { res.Flush() })
	assert.False(t, res.Committed)
	assert.False(t, beforeCalled)
}

func TestHeadResponseWriter_WriteHeader_Informational(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	res := NewResponse(rec, e.Logger)
	w := &headResponseWriter{rw: res}

	w.WriteHeader(http.StatusEarlyHints)
	assert.False(t, w.wroteStatus)
	assert.False(t, res.Committed)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("ok"))
	w.commit()

	assert.Equal(t, http.StatusCreated, res.Status)
	assert.Equal(t, "2", rec.Header().Get(HeaderContentLength))
}

func TestDelayedStatusWriter_Flush_SendsDelayedStatus(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	res := NewResponse(rec, e.Logger)
	w := &delayedStatusWriter{ResponseWriter: res, status: http.StatusCreated}

	w.WriteHeader(http.StatusEarlyHints)
	assert.False(t, w.committed)
	w.Flush()
	_, _ = w.Write([]byte("ok"))

	// httptest.ResponseRecorder keeps the first status it sees, including 1xx, so check Response.
	assert.Equal(t, http.StatusCreated, res.Status)
	assert.True(t, res.Committed)
	assert.True(t, rec.Flushed)
}

func TestDelayedStatusWriter_Flush_PanicsBeforeCommit(t *testing.T) {
	e := New()
	res := NewResponse(new(testResponseWriter), e.Logger)
	w := &delayedStatusWriter{ResponseWriter: res, status: http.StatusCreated}

	assert.Panics(t, func() { w.Flush() })
	assert.False(t, w.committed)
	assert.False(t, res.Committed)
}
