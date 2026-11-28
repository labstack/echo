// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jsonpOuter struct{}
type jsonpMiddle struct{}

type jsonpDelegatingSerializer struct{ DefaultJSONSerializer }

func (s jsonpDelegatingSerializer) Serialize(c *Context, v any, indent string) error {
	switch v.(type) {
	case jsonpOuter:
		return c.JSONP(http.StatusAccepted, "callback", jsonpMiddle{})
	case jsonpMiddle:
		return c.JSON(http.StatusCreated, map[string]bool{"ok": true})
	}
	// Detect a writer cycle before encoding can recurse through Header or Write.
	seen := make(map[http.ResponseWriter]bool)
	for w := c.Response(); ; {
		if seen[w] {
			return errors.New("response writer cycle")
		}
		seen[w] = true
		unwrapper, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			break
		}
		w = unwrapper.Unwrap()
	}
	return s.DefaultJSONSerializer.Serialize(c, v, indent)
}

func TestJSONPDelegatingSerializer(t *testing.T) {
	for _, contentType := range []string{"", "application/custom-javascript"} {
		t.Run(contentType, func(t *testing.T) {
			e := New()
			e.JSONSerializer = jsonpDelegatingSerializer{}
			rec := httptest.NewRecorder()
			c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
			if contentType != "" {
				c.Response().Header().Set(HeaderContentType, contentType)
			}
			original := c.Response()
			require.NoError(t, c.JSONP(http.StatusAccepted, "callback", jsonpMiddle{}))
			assert.Same(t, original, c.Response())
			assert.Equal(t, http.StatusAccepted, rec.Code)
			assert.Equal(t, "callback({\"ok\":true}\n);", rec.Body.String())
			if contentType == "" {
				contentType = MIMEApplicationJavaScriptCharsetUTF8
			}
			assert.Equal(t, contentType, rec.Result().Header.Get(HeaderContentType))
		})
	}
}

func TestJSONPInsideJSONDelegatingSerializer(t *testing.T) {
	e := New()
	e.JSONSerializer = jsonpDelegatingSerializer{}
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
	original := c.Response()
	require.NoError(t, c.JSON(http.StatusAccepted, jsonpOuter{}))
	assert.Same(t, original, c.Response())
	assert.Equal(t, "callback({\"ok\":true}\n);", rec.Body.String())
}

type jsonpFlushErrorWriter struct {
	*httptest.ResponseRecorder
	err error
}

func (w *jsonpFlushErrorWriter) FlushError() error { return w.err }

type jsonpNoFlushWriter struct{ http.ResponseWriter }

func TestJSONPFlushErrors(t *testing.T) {
	flushError := errors.New("flush failed")
	for _, tc := range []struct {
		name       string
		makeWriter func(*httptest.ResponseRecorder) http.ResponseWriter
		want       error
		body       string
	}{
		{"underlying error", func(rec *httptest.ResponseRecorder) http.ResponseWriter {
			return &jsonpFlushErrorWriter{ResponseRecorder: rec, err: flushError}
		}, flushError, "callback("},
		{"unsupported", func(rec *httptest.ResponseRecorder) http.ResponseWriter {
			return jsonpNoFlushWriter{rec}
		}, http.ErrNotSupported, ""},
		{"prefix error", func(rec *httptest.ResponseRecorder) http.ResponseWriter {
			return &jsonpFailWriter{ResponseRecorder: rec, failAt: 1, err: flushError}
		}, flushError, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.JSONSerializer = jsonpTestSerializer{serialize: func(c *Context) error {
				return http.NewResponseController(c.Response()).Flush()
			}}
			rec := httptest.NewRecorder()
			c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
			c.SetResponse(tc.makeWriter(rec))
			require.ErrorIs(t, c.JSONP(http.StatusAccepted, "callback", nil), tc.want)
			assert.Equal(t, tc.body, rec.Body.String())
		})
	}
}

func TestJSONPSerializerCannotReplaceStatus(t *testing.T) {
	for _, status := range []int{http.StatusCreated, http.StatusEarlyHints, http.StatusSwitchingProtocols} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			e := New()
			e.JSONSerializer = jsonpTestSerializer{serialize: func(c *Context) error {
				c.Response().WriteHeader(status)
				_, err := c.Response().Write([]byte("null"))
				return err
			}}
			rec := httptest.NewRecorder()
			c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
			require.NoError(t, c.JSONP(http.StatusAccepted, "callback", nil))
			assert.Equal(t, http.StatusAccepted, rec.Code)
			assert.Equal(t, "callback(null);", rec.Body.String())
		})
	}
}
