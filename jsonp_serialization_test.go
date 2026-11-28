// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jsonpTestSerializer struct {
	DefaultJSONSerializer
	serialize func(*Context) error
}

func (s jsonpTestSerializer) Serialize(c *Context, _ any, _ string) error {
	return s.serialize(c)
}

func TestJSONPErrorBeforeSerializationWrite(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(map[bool]string{false: "default serializer", true: "custom serializer"}[custom], func(t *testing.T) {
			e := New()
			if custom {
				e.JSONSerializer = jsonpTestSerializer{serialize: func(*Context) error {
					return errors.New("cannot serialize")
				}}
				e.HTTPErrorHandler = func(c *Context, _ error) {
					assert.NoError(t, c.String(http.StatusInternalServerError, "serialization failed"))
				}
			}
			e.GET("/", func(c *Context) error {
				return c.JSONP(http.StatusAccepted, "callback", make(chan bool))
			})
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotContains(t, rec.Body.String(), "callback(")
			assert.Empty(t, rec.Header().Get(HeaderXContentTypeOptions))
			if custom {
				assert.Equal(t, MIMETextPlainCharsetUTF8, rec.Header().Get(HeaderContentType))
			} else {
				assert.Equal(t, MIMEApplicationJSON, rec.Header().Get(HeaderContentType))
				assert.JSONEq(t, `{"message":"Internal Server Error"}`, rec.Body.String())
			}
		})
	}
}

func TestJSONPSerializerResponseControls(t *testing.T) {
	for _, tc := range []struct {
		name      string
		serialize func(*Context) error
		body      string
	}{
		{"empty success", func(*Context) error { return nil }, "callback();"},
		{"chunked success", func(c *Context) error {
			_, err := c.Response().Write([]byte(`{"ok":`))
			if err != nil {
				return err
			}
			_, err = c.Response().Write([]byte(`true}`))
			return err
		}, `callback({"ok":true});`},
		{"error after write", func(c *Context) error {
			_, err := c.Response().Write([]byte(`{"partial":`))
			if err != nil {
				return err
			}
			return errors.New("after write")
		}, `callback({"partial":`},
		{"flush before write", func(c *Context) error {
			if err := http.NewResponseController(c.Response()).Flush(); err != nil {
				return err
			}
			_, err := c.Response().Write([]byte(`null`))
			return err
		}, `callback(null);`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.JSONSerializer = jsonpTestSerializer{serialize: tc.serialize}
			e.GET("/", func(c *Context) error { return c.JSONP(http.StatusAccepted, "callback", nil) })
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Equal(t, http.StatusAccepted, rec.Code)
			assert.Equal(t, tc.body, rec.Body.String())
			assert.Equal(t, MIMEApplicationJavaScriptCharsetUTF8, rec.Result().Header.Get(HeaderContentType))
			assert.Equal(t, "nosniff", rec.Result().Header.Get(HeaderXContentTypeOptions))
		})
	}
}

func TestJSONPInvalidCallbackDoesNotSerialize(t *testing.T) {
	e := New()
	e.JSONSerializer = jsonpTestSerializer{serialize: func(*Context) error {
		t.Fatal("invalid callback must be rejected before serialization")
		return nil
	}}
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
	require.ErrorIs(t, c.JSONP(http.StatusAccepted, "callback()", nil), ErrInvalidJSONPCallback)
	assert.Empty(t, rec.Body.String())
}

type jsonpFailWriter struct {
	*httptest.ResponseRecorder
	failAt int
	writes int
	err    error
}

func (w *jsonpFailWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.failAt {
		return 0, w.err
	}
	return w.ResponseRecorder.Write(p)
}

func TestJSONPWriterErrors(t *testing.T) {
	writeError := errors.New("write failed")
	for _, tc := range []struct {
		name   string
		failAt int
		err    error
		want   error
	}{
		{"callback", 1, writeError, writeError},
		{"opening parenthesis", 2, writeError, writeError},
		{"payload", 3, writeError, writeError},
		{"closing parenthesis", 4, writeError, writeError},
		{"short callback", 1, nil, io.ErrShortWrite},
		{"short opening parenthesis", 2, nil, io.ErrShortWrite},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			writer := &jsonpFailWriter{ResponseRecorder: httptest.NewRecorder(), failAt: tc.failAt, err: tc.err}
			c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), writer)
			original := c.Response()
			require.ErrorIs(t, c.JSONP(http.StatusAccepted, "callback", nil), tc.want)
			assert.Same(t, original, c.Response())
		})
	}
}
