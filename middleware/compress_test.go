// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzip_NoAcceptEncodingHeader(t *testing.T) {
	// Skip if no Accept-Encoding header
	h := Gzip()(func(c *echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	assert.NoError(t, err)

	assert.Equal(t, "test", rec.Body.String())
}

func TestMustGzipWithConfig_panics(t *testing.T) {
	assert.Panics(t, func() {
		GzipWithConfig(GzipConfig{Level: 999})
	})
}

func TestGzip_AcceptEncodingHeader(t *testing.T) {
	h := Gzip()(func(c *echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	assert.NoError(t, err)

	assert.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), echo.MIMETextPlain)

	r, err := gzip.NewReader(rec.Body)
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	defer r.Close()
	buf.ReadFrom(r)
	assert.Equal(t, "test", buf.String())
}

func TestGzip_AcceptEncodingQuality(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		gzip    bool
	}{
		{"explicit refusal", []string{"gzip;q=0"}, false},
		{"zero fraction", []string{"gzip;q=0.000"}, false},
		{"partial token", []string{"custom-gzip"}, false},
		{"uppercase", []string{"GZIP; Q=1.000"}, true},
		{"minimum positive weight", []string{"gzip;q=0.001"}, true},
		{"multiple fields", []string{"br", "gzip;q=0.5"}, true},
		{"wildcard", []string{"br, *;q=0.5"}, true},
		{"wildcard refusal", []string{"*;q=0"}, false},
		{"explicit refusal overrides wildcard", []string{"*", "gzip;q=0"}, false},
		{"explicit allowance overrides wildcard", []string{"*;q=0, gzip"}, true},
		{"empty", []string{""}, false},
		{"identity", []string{"identity"}, false},
		{"x-gzip", []string{"x-gzip"}, true},
		{"x-gzip case-insensitive", []string{"X-GZIP"}, true},
		{"x-gzip refusal", []string{"x-gzip;q=0"}, false},
		{"invalid weight", []string{"gzip;q=bogus"}, false},
		{"out-of-range weight", []string{"gzip;q=2"}, false},
		{"too much precision", []string{"gzip;q=0.0001"}, false},
		{"nonzero digit after one", []string{"gzip;q=1.1"}, false},
		{"exponent", []string{"gzip;q=1e-1"}, false},
		{"malformed explicit weight overrides wildcard", []string{"*, gzip;q="}, false},
		{"duplicate weights", []string{"gzip;q=0;q=1"}, false},
		{"duplicate refusal then allowance", []string{"gzip;q=0", "gzip;q=1"}, false},
		{"duplicate allowance then refusal", []string{"gzip;q=1, gzip;q=0"}, false},
		{"duplicate wildcard refusal", []string{"*;q=0, *"}, false},
	}

	e := echo.New()
	e.Use(Gzip())
	e.GET("/", func(c *echo.Context) error { return c.String(http.StatusOK, "test") })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for _, value := range tt.headers {
				req.Header.Add(echo.HeaderAcceptEncoding, value)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, echo.HeaderAcceptEncoding, rec.Header().Get(echo.HeaderVary))
			if !tt.gzip {
				assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
				assert.Equal(t, "test", rec.Body.String())
				return
			}
			require.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
			reader, err := gzip.NewReader(rec.Body)
			require.NoError(t, err)
			defer reader.Close()
			body, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, "test", string(body))
		})
	}
}

func TestGzip_CompressesIdentityContentEncoding(t *testing.T) {
	for _, value := range []string{"identity", "IDENTITY", "identity, identity"} {
		t.Run(value, func(t *testing.T) {
			e := echo.New()
			e.Use(Gzip())
			e.GET("/", func(c *echo.Context) error {
				c.Response().Header().Set(echo.HeaderContentEncoding, value)
				return c.String(http.StatusOK, "test")
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
			reader, err := gzip.NewReader(rec.Body)
			require.NoError(t, err)
			defer reader.Close()
			body, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, "test", string(body))
		})
	}
}

func TestGzip_DropsContentLengthOnImplicitHeader(t *testing.T) {
	for _, flush := range []bool{false, true} {
		t.Run(fmt.Sprintf("flush=%v", flush), func(t *testing.T) {
			e := echo.New()
			e.Use(Gzip())
			e.GET("/", func(c *echo.Context) error {
				w := c.Response()
				w.Header().Set(echo.HeaderContentLength, "4")
				if _, err := w.Write([]byte("test")); err != nil {
					return err
				}
				if flush {
					return http.NewResponseController(w).Flush()
				}
				return nil
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			require.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
			assert.Empty(t, rec.Header().Get(echo.HeaderContentLength))
			reader, err := gzip.NewReader(rec.Body)
			require.NoError(t, err)
			defer reader.Close()
			body, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, "test", string(body))
		})
	}
}

func TestGzip_PreservesContentLengthWhenNotCompressed(t *testing.T) {
	encoded, err := gzipString("already encoded")
	require.NoError(t, err)

	tests := []struct {
		name      string
		minLength int
		handler   echo.HandlerFunc
		body      []byte
	}{
		{
			name:      "below min length",
			minLength: 100,
			handler: func(c *echo.Context) error {
				w := c.Response()
				w.Header().Set(echo.HeaderContentLength, "4")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("test"))
				return err
			},
			body: []byte("test"),
		},
		{
			name: "encoding set after WriteHeader",
			handler: func(c *echo.Context) error {
				w := c.Response()
				w.Header().Set(echo.HeaderContentLength, strconv.Itoa(len(encoded)))
				w.WriteHeader(http.StatusOK)
				w.Header().Set(echo.HeaderContentEncoding, gzipScheme)
				_, err := w.Write(encoded)
				return err
			},
			body: encoded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Use(GzipWithConfig(GzipConfig{MinLength: tt.minLength}))
			e.GET("/", tt.handler)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, strconv.Itoa(len(tt.body)), rec.Header().Get(echo.HeaderContentLength))
			assert.Equal(t, tt.body, rec.Body.Bytes())
		})
	}
}

func TestGzip_ErrorClearsEncodingHeaders(t *testing.T) {
	tests := []struct {
		name        string
		encoding    string
		writeHeader bool
		code        int
	}{
		{name: "gzip", encoding: gzipScheme, code: http.StatusBadRequest},
		{name: "case-insensitive gzip", encoding: "GZIP", code: http.StatusBadRequest},
		{name: "other coding", encoding: "br", code: http.StatusBadRequest},
		{name: "after WriteHeader", writeHeader: true, code: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Use(Gzip())
			e.GET("/", func(c *echo.Context) error {
				w := c.Response()
				if tt.encoding != "" {
					w.Header().Set(echo.HeaderContentEncoding, tt.encoding)
				}
				w.Header().Set(echo.HeaderContentLength, "50")
				if tt.writeHeader {
					w.WriteHeader(http.StatusOK)
				}
				return echo.ErrBadRequest
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.code, rec.Code)
			assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
			assert.Empty(t, rec.Header().Get(echo.HeaderContentLength))
		})
	}
}

func TestGzip_WriteReportsCallerBytes(t *testing.T) {
	e := echo.New()
	e.Use(GzipWithConfig(GzipConfig{MinLength: 10}))
	e.GET("/", func(c *echo.Context) error {
		_, err := io.Copy(c.Response(), iotest.OneByteReader(strings.NewReader(strings.Repeat("a", 20))))
		return err
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
	reader, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer reader.Close()
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat("a", 20), string(body))
}

func TestGzip_KeepsContentLengthWithoutBody(t *testing.T) {
	e := echo.New()
	e.Use(Gzip())
	e.HEAD("/", func(c *echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentLength, "1234")
		return c.NoContent(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodHead, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "1234", rec.Header().Get(echo.HeaderContentLength))
	assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
}

func TestGzip_PassthroughFlushCommitsResponse(t *testing.T) {
	e := echo.New()
	e.Use(Gzip())
	e.GET("/", func(c *echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentEncoding, gzipScheme)
		if err := http.NewResponseController(c.Response()).Flush(); err != nil {
			return err
		}
		r, err := echo.UnwrapResponse(c.Response())
		require.NoError(t, err)
		assert.True(t, r.Committed)
		return echo.ErrBadRequest
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestGzip_ForwardsInformationalStatus(t *testing.T) {
	e := echo.New()
	e.Use(Gzip())
	e.GET("/", func(c *echo.Context) error {
		c.Response().Header().Set("Link", "</style.css>; rel=preload; as=style")
		c.Response().WriteHeader(http.StatusEarlyHints)
		return c.String(http.StatusCreated, "ok")
	})
	srv := httptest.NewServer(e)
	defer srv.Close()

	var informational []int
	trace := &httptrace.ClientTrace{
		Got1xxResponse: func(code int, _ textproto.MIMEHeader) error {
			informational = append(informational, code)
			return nil
		},
	}
	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(context.Background(), trace), http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	tr := &http.Transport{DisableCompression: true}
	defer tr.CloseIdleConnections()
	res, err := tr.RoundTrip(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, []int{http.StatusEarlyHints}, informational)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	reader, err := gzip.NewReader(res.Body)
	require.NoError(t, err)
	defer reader.Close()
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "ok", string(body))
}

func TestGzip_FlushWithoutWriteProducesValidStream(t *testing.T) {
	e := echo.New()
	e.Use(Gzip())
	e.GET("/", func(c *echo.Context) error {
		return http.NewResponseController(c.Response()).Flush()
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
	reader, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer reader.Close()
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Empty(t, body)
}

func TestGzip_PreservesEncodedResponseOnWriteOrFlush(t *testing.T) {
	encoded, err := gzipString("already encoded")
	require.NoError(t, err)

	for _, flush := range []bool{false, true} {
		name := "implicit header"
		if flush {
			name = "flush before write"
		}
		t.Run(name, func(t *testing.T) {
			e := echo.New()
			e.Use(Gzip())
			e.GET("/", func(c *echo.Context) error {
				w := c.Response()
				w.Header().Set(echo.HeaderContentEncoding, gzipScheme)
				if flush {
					if err := http.NewResponseController(w).Flush(); err != nil {
						return err
					}
				}
				_, err := w.Write(encoded)
				return err
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, gzipScheme, rec.Result().Header.Get(echo.HeaderContentEncoding))
			assert.Equal(t, flush, rec.Flushed)
			assert.Equal(t, encoded, rec.Body.Bytes())
		})
	}
}

func TestGzip_PreservesImplicitHeadEncoding(t *testing.T) {
	e := echo.New()
	e.Use(Gzip())
	e.HEAD("/", func(c *echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentEncoding, gzipScheme)
		c.Response().Header().Set(echo.HeaderContentLength, "42")
		return nil
	})

	req := httptest.NewRequest(http.MethodHead, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, gzipScheme, rec.Result().Header.Get(echo.HeaderContentEncoding))
	assert.Equal(t, int64(42), rec.Result().ContentLength)
	assert.Empty(t, rec.Body.Bytes())
}

func TestGzip_chunked(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	chunkChan := make(chan struct{})
	waitChan := make(chan struct{})
	h := Gzip()(func(c *echo.Context) error {
		rc := http.NewResponseController(c.Response())
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Transfer-Encoding", "chunked")

		// Write and flush the first part of the data
		c.Response().Write([]byte("first\n"))
		rc.Flush()

		chunkChan <- struct{}{}
		<-waitChan

		// Write and flush the second part of the data
		c.Response().Write([]byte("second\n"))
		rc.Flush()

		chunkChan <- struct{}{}
		<-waitChan

		// Write the final part of the data and return
		c.Response().Write([]byte("third"))

		chunkChan <- struct{}{}
		return nil
	})

	go func() {
		err := h(c)
		chunkChan <- struct{}{}
		assert.NoError(t, err)
	}()

	<-chunkChan // wait for first write
	waitChan <- struct{}{}

	<-chunkChan // wait for second write
	waitChan <- struct{}{}

	<-chunkChan                      // wait for final write in handler
	<-chunkChan                      // wait for return from handler
	time.Sleep(5 * time.Millisecond) // to have time for flushing

	assert.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))

	r, err := gzip.NewReader(rec.Body)
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	assert.Equal(t, "first\nsecond\nthird", buf.String())
}

func TestGzip_NoContent(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Gzip()(func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	if assert.NoError(t, h(c)) {
		assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func TestGzip_Empty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Gzip()(func(c *echo.Context) error {
		return c.String(http.StatusOK, "")
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
		assert.Equal(t, "text/plain; charset=UTF-8", rec.Header().Get(echo.HeaderContentType))
		r, err := gzip.NewReader(rec.Body)
		if assert.NoError(t, err) {
			var buf bytes.Buffer
			buf.ReadFrom(r)
			assert.Equal(t, "", buf.String())
		}
	}
}

func TestGzip_ErrorReturned(t *testing.T) {
	e := echo.New()
	e.Use(Gzip())
	e.GET("/", func(c *echo.Context) error {
		return echo.ErrNotFound
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
}

func TestGzipWithConfig_invalidLevel(t *testing.T) {
	mw, err := GzipConfig{Level: 12}.ToMiddleware()
	assert.EqualError(t, err, "invalid gzip level")
	assert.Nil(t, mw)
}

// Issue #806
func TestGzipWithStatic(t *testing.T) {
	e := echo.New()
	e.Filesystem = os.DirFS("../")

	e.Use(Gzip())
	e.Static("/test", "_fixture/images")
	req := httptest.NewRequest(http.MethodGet, "/test/walle.png", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Data is written out in chunks when Content-Length == "", so only
	// validate the content length if it's not set.
	if cl := rec.Header().Get("Content-Length"); cl != "" {
		assert.Equal(t, cl, rec.Body.Len())
	}
	r, err := gzip.NewReader(rec.Body)
	if assert.NoError(t, err) {
		defer r.Close()
		want, err := os.ReadFile("../_fixture/images/walle.png")
		if assert.NoError(t, err) {
			buf := new(bytes.Buffer)
			buf.ReadFrom(r)
			assert.Equal(t, want, buf.Bytes())
		}
	}
}

func TestGzipWithMinLength(t *testing.T) {
	e := echo.New()
	// Minimal response length
	e.Use(GzipWithConfig(GzipConfig{MinLength: 10}))
	e.GET("/", func(c *echo.Context) error {
		c.Response().Write([]byte("foobarfoobar"))
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))
	r, err := gzip.NewReader(rec.Body)
	if assert.NoError(t, err) {
		buf := new(bytes.Buffer)
		defer r.Close()
		buf.ReadFrom(r)
		assert.Equal(t, "foobarfoobar", buf.String())
	}
}

func TestGzipWithMinLengthTooShort(t *testing.T) {
	e := echo.New()
	// Minimal response length
	e.Use(GzipWithConfig(GzipConfig{MinLength: 10}))
	e.GET("/", func(c *echo.Context) error {
		c.Response().Write([]byte("test"))
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "", rec.Header().Get(echo.HeaderContentEncoding))
	assert.Contains(t, rec.Body.String(), "test")
}

func TestGzipWithResponseWithoutBody(t *testing.T) {
	e := echo.New()

	e.Use(Gzip())
	e.GET("/", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "http://localhost")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "", rec.Header().Get(echo.HeaderContentEncoding))
}

func TestGzipWithMinLengthChunked(t *testing.T) {
	e := echo.New()

	// Gzip chunked
	chunkBuf := make([]byte, 5)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()

	var r *gzip.Reader = nil

	c := e.NewContext(req, rec)
	next := func(c *echo.Context) error {
		rc := http.NewResponseController(c.Response())
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Transfer-Encoding", "chunked")

		// Write and flush the first part of the data
		c.Response().Write([]byte("test\n"))
		rc.Flush()

		// Read the first part of the data
		assert.True(t, rec.Flushed)
		assert.Equal(t, gzipScheme, rec.Header().Get(echo.HeaderContentEncoding))

		var err error
		r, err = gzip.NewReader(rec.Body)
		assert.NoError(t, err)

		_, err = io.ReadFull(r, chunkBuf)
		assert.NoError(t, err)
		assert.Equal(t, "test\n", string(chunkBuf))

		// Write and flush the second part of the data
		c.Response().Write([]byte("test\n"))
		rc.Flush()

		_, err = io.ReadFull(r, chunkBuf)
		assert.NoError(t, err)
		assert.Equal(t, "test\n", string(chunkBuf))

		// Write the final part of the data and return
		c.Response().Write([]byte("test"))
		return nil
	}
	err := GzipWithConfig(GzipConfig{MinLength: 10})(next)(c)

	assert.NoError(t, err)
	assert.NotNil(t, r)

	buf := new(bytes.Buffer)

	buf.ReadFrom(r)
	assert.Equal(t, "test", buf.String())

	r.Close()
}

func TestGzipWithMinLengthNoContent(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := GzipWithConfig(GzipConfig{MinLength: 10})(func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	if assert.NoError(t, h(c)) {
		assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(echo.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func TestGzipResponseWriter_CanUnwrap(t *testing.T) {
	trwu := &testResponseWriterUnwrapper{rw: httptest.NewRecorder()}
	bdrw := gzipResponseWriter{
		ResponseWriter: trwu,
	}
	result := bdrw.Unwrap()
	assert.Equal(t, trwu, result)
}

func TestGzipResponseWriter_CanHijack(t *testing.T) {
	trwu := testResponseWriterUnwrapperHijack{testResponseWriterUnwrapper: testResponseWriterUnwrapper{rw: httptest.NewRecorder()}}
	bdrw := gzipResponseWriter{
		ResponseWriter: &trwu, // this RW supports hijacking through unwrapping
	}
	_, _, err := bdrw.Hijack()
	assert.EqualError(t, err, "can hijack")
}

func TestGzipResponseWriter_CanNotHijack(t *testing.T) {
	trwu := testResponseWriterUnwrapper{rw: httptest.NewRecorder()}
	bdrw := gzipResponseWriter{
		ResponseWriter: &trwu, // this RW supports hijacking through unwrapping
	}
	_, _, err := bdrw.Hijack()
	assert.EqualError(t, err, "feature not supported")
}

func BenchmarkGzip(b *testing.B) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, gzipScheme)

	h := Gzip()(func(c *echo.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Gzip
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h(c)
	}
}
