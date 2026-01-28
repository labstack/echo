// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
)

type Template struct {
	templates *template.Template
}

var testUser = user{ID: 1, Name: "Jon Snow"}

func BenchmarkAllocJSONP(b *testing.B) {
	e := New()
	e.Logger = slog.New(slog.DiscardHandler)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.JSONP(http.StatusOK, "callback", testUser)
	}
}

func BenchmarkAllocJSON(b *testing.B) {
	e := New()
	e.Logger = slog.New(slog.DiscardHandler)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.JSON(http.StatusOK, testUser)
	}
}

func BenchmarkAllocXML(b *testing.B) {
	e := New()
	e.Logger = slog.New(slog.DiscardHandler)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.XML(http.StatusOK, testUser)
	}
}

func BenchmarkRealIPForHeaderXForwardFor(b *testing.B) {
	c := Context{request: &http.Request{
		Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1, 127.0.1.1, "}},
	}}
	for i := 0; i < b.N; i++ {
		c.RealIP()
	}
}

func (t *Template) Render(c *Context, w io.Writer, name string, data any) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func TestContextEcho(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	assert.Equal(t, e, c.Echo())
}

func TestContextRequest(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	assert.NotNil(t, c.Request())
	assert.Equal(t, req, c.Request())
}

func TestContextResponse(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	assert.NotNil(t, c.Response())
}

func TestContextRenderTemplate(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	tmpl := &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	c.Echo().Renderer = tmpl
	err := c.Render(http.StatusOK, "hello", "Jon Snow")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello, Jon Snow!", rec.Body.String())
	}
}

func TestContextRenderTemplateError(t *testing.T) {
	// we test that when template rendering fails, no response is sent to the client yet, so the global error handler can decide what to do
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tmpl := &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	c.Echo().Renderer = tmpl
	err := c.Render(http.StatusOK, "not_existing", "Jon Snow")

	assert.EqualError(t, err, `template: no template "not_existing" associated with template "hello"`)
	assert.Equal(t, http.StatusOK, rec.Code) // status code must not be sent to the client
	assert.Empty(t, rec.Body.String())       // body must not be sent to the client
}

func TestContextRenderErrorsOnNoRenderer(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.Echo().Renderer = nil
	assert.Error(t, c.Render(http.StatusOK, "hello", "Jon Snow"))
}

func TestContextStream(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		for i := 0; i < 3; i++ {
			fmt.Fprintf(w, "data: index %v\n\n", i)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	err := c.Stream(http.StatusOK, "text/event-stream", r)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/event-stream", rec.Header().Get(HeaderContentType))
		assert.Equal(t, "data: index 0\n\ndata: index 1\n\ndata: index 2\n\n", rec.Body.String())
	}
}

func TestContextHTML(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := NewContext(req, rec)

	err := c.HTML(http.StatusOK, "Hi, Jon Snow")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMETextHTMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, "Hi, Jon Snow", rec.Body.String())
	}
}

func TestContextHTMLBlob(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := NewContext(req, rec)

	err := c.HTMLBlob(http.StatusOK, []byte("Hi, Jon Snow"))
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMETextHTMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, "Hi, Jon Snow", rec.Body.String())
	}
}

func TestContextJSON(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	c := e.NewContext(req, rec)

	err := c.JSON(http.StatusOK, user{ID: 1, Name: "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJSON, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSON+"\n", rec.Body.String())
	}
}

func TestContextJSONErrorsOut(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	c := e.NewContext(req, rec)

	err := c.JSON(http.StatusOK, make(chan bool))
	assert.EqualError(t, err, "json: unsupported type: chan bool")

	assert.Equal(t, http.StatusOK, rec.Code) // status code must not be sent to the client
	assert.Empty(t, rec.Body.String())       // body must not be sent to the client
}

func TestContextJSONWithNotEchoResponse(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	c := e.NewContext(req, rec)

	c.SetResponse(rec)

	err := c.JSON(http.StatusCreated, map[string]float64{"foo": math.NaN()})
	assert.EqualError(t, err, "json: unsupported value: NaN")

	assert.Equal(t, http.StatusOK, rec.Code) // status code must not be sent to the client
	assert.Empty(t, rec.Body.String())       // body must not be sent to the client
}

func TestContextJSONPretty(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)

	err := c.JSONPretty(http.StatusOK, user{ID: 1, Name: "Jon Snow"}, "  ")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJSON, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSONPretty+"\n", rec.Body.String())
	}
}

func TestContextJSONWithEmptyIntent(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)

	u := user{ID: 1, Name: "Jon Snow"}
	emptyIndent := ""
	buf := new(bytes.Buffer)

	enc := json.NewEncoder(buf)
	enc.SetIndent(emptyIndent, emptyIndent)
	_ = enc.Encode(u)
	err := c.JSONPretty(http.StatusOK, user{ID: 1, Name: "Jon Snow"}, emptyIndent)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJSON, rec.Header().Get(HeaderContentType))
		assert.Equal(t, buf.String(), rec.Body.String())
	}
}

func TestContextJSONP(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)

	callback := "callback"
	err := c.JSONP(http.StatusOK, callback, user{ID: 1, Name: "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJavaScriptCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, callback+"("+userJSON+"\n);", rec.Body.String())
	}
}

func TestContextJSONBlob(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)

	data, err := json.Marshal(user{ID: 1, Name: "Jon Snow"})
	assert.NoError(t, err)
	err = c.JSONBlob(http.StatusOK, data)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJSON, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSON, rec.Body.String())
	}
}

func TestContextJSONPBlob(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)

	callback := "callback"
	data, err := json.Marshal(user{ID: 1, Name: "Jon Snow"})
	assert.NoError(t, err)
	err = c.JSONPBlob(http.StatusOK, callback, data)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJavaScriptCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, callback+"("+userJSON+");", rec.Body.String())
	}
}

func TestContextXML(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)

	err := c.XML(http.StatusOK, user{ID: 1, Name: "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+userXML, rec.Body.String())
	}
}

func TestContextXMLPretty(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)

	err := c.XMLPretty(http.StatusOK, user{ID: 1, Name: "Jon Snow"}, "  ")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+userXMLPretty, rec.Body.String())
	}
}

func TestContextXMLBlob(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)

	data, err := xml.Marshal(user{ID: 1, Name: "Jon Snow"})
	assert.NoError(t, err)
	err = c.XMLBlob(http.StatusOK, data)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+userXML, rec.Body.String())
	}
}

func TestContextXMLWithEmptyIntent(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)

	u := user{ID: 1, Name: "Jon Snow"}
	emptyIndent := ""
	buf := new(bytes.Buffer)

	enc := xml.NewEncoder(buf)
	enc.Indent(emptyIndent, emptyIndent)
	_ = enc.Encode(u)
	err := c.XMLPretty(http.StatusOK, user{ID: 1, Name: "Jon Snow"}, emptyIndent)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+buf.String(), rec.Body.String())
	}
}

func TestContext_JSON_CommitsCustomResponseCode(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := c.JSON(http.StatusCreated, user{ID: 1, Name: "Jon Snow"})

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, MIMEApplicationJSON, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSON+"\n", rec.Body.String())
	}
}

func TestContextAttachment(t *testing.T) {
	var testCases = []struct {
		name         string
		whenName     string
		expectHeader string
	}{
		{
			name:         "ok",
			whenName:     "walle.png",
			expectHeader: `attachment; filename="walle.png"`,
		},
		{
			name:         "ok, escape quotes in malicious filename",
			whenName:     `malicious.sh"; \"; dummy=.txt`,
			expectHeader: `attachment; filename="malicious.sh\"; \\\"; dummy=.txt"`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			c := e.NewContext(req, rec)

			err := c.Attachment("_fixture/images/walle.png", tc.whenName)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectHeader, rec.Header().Get(HeaderContentDisposition))

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, 219885, rec.Body.Len())
			}
		})
	}
}

func TestContextInline(t *testing.T) {
	var testCases = []struct {
		name         string
		whenName     string
		expectHeader string
	}{
		{
			name:         "ok",
			whenName:     "walle.png",
			expectHeader: `inline; filename="walle.png"`,
		},
		{
			name:         "ok, escape quotes in malicious filename",
			whenName:     `malicious.sh"; \"; dummy=.txt`,
			expectHeader: `inline; filename="malicious.sh\"; \\\"; dummy=.txt"`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			c := e.NewContext(req, rec)

			err := c.Inline("_fixture/images/walle.png", tc.whenName)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectHeader, rec.Header().Get(HeaderContentDisposition))

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, 219885, rec.Body.Len())
			}
		})
	}
}

func TestContextNoContent(t *testing.T) {
	e := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?pretty", nil)
	c := e.NewContext(req, rec)

	c.NoContent(http.StatusOK)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestContextCookie(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	theme := "theme=light"
	user := "user=Jon Snow"
	req.Header.Add(HeaderCookie, theme)
	req.Header.Add(HeaderCookie, user)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Read single
	cookie, err := c.Cookie("theme")
	if assert.NoError(t, err) {
		assert.Equal(t, "theme", cookie.Name)
		assert.Equal(t, "light", cookie.Value)
	}

	// Read multiple
	for _, cookie := range c.Cookies() {
		switch cookie.Name {
		case "theme":
			assert.Equal(t, "light", cookie.Value)
		case "user":
			assert.Equal(t, "Jon Snow", cookie.Value)
		}
	}

	// Write
	cookie = &http.Cookie{
		Name:     "SSID",
		Value:    "Ap4PGTEq",
		Domain:   "labstack.com",
		Path:     "/",
		Expires:  time.Now(),
		Secure:   true,
		HttpOnly: true,
	}
	c.SetCookie(cookie)
	assert.Contains(t, rec.Header().Get(HeaderSetCookie), "SSID")
	assert.Contains(t, rec.Header().Get(HeaderSetCookie), "Ap4PGTEq")
	assert.Contains(t, rec.Header().Get(HeaderSetCookie), "labstack.com")
	assert.Contains(t, rec.Header().Get(HeaderSetCookie), "Secure")
	assert.Contains(t, rec.Header().Get(HeaderSetCookie), "HttpOnly")
}

func TestContext_PathValues(t *testing.T) {
	var testCases = []struct {
		name   string
		given  PathValues
		expect PathValues
	}{
		{
			name: "param exists",
			given: PathValues{
				{Name: "uid", Value: "101"},
				{Name: "fid", Value: "501"},
			},
			expect: PathValues{
				{Name: "uid", Value: "101"},
				{Name: "fid", Value: "501"},
			},
		},
		{
			name:   "params is empty",
			given:  PathValues{},
			expect: PathValues{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			c := e.NewContext(req, nil)

			c.SetPathValues(tc.given)

			assert.EqualValues(t, tc.expect, c.PathValues())
		})
	}
}

func TestContext_PathParam(t *testing.T) {
	var testCases = []struct {
		name          string
		given         PathValues
		whenParamName string
		expect        string
	}{
		{
			name: "param exists",
			given: PathValues{
				{Name: "uid", Value: "101"},
				{Name: "fid", Value: "501"},
			},
			whenParamName: "uid",
			expect:        "101",
		},
		{
			name: "multiple same param values exists - return first",
			given: PathValues{
				{Name: "uid", Value: "101"},
				{Name: "uid", Value: "202"},
				{Name: "fid", Value: "501"},
			},
			whenParamName: "uid",
			expect:        "101",
		},
		{
			name: "param does not exists",
			given: PathValues{
				{Name: "uid", Value: "101"},
			},
			whenParamName: "nope",
			expect:        "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			c := e.NewContext(req, nil)

			c.SetPathValues(tc.given)

			assert.EqualValues(t, tc.expect, c.Param(tc.whenParamName))
		})
	}
}

func TestContext_PathParamDefault(t *testing.T) {
	var testCases = []struct {
		name             string
		given            PathValues
		whenParamName    string
		whenDefaultValue string
		expect           string
	}{
		{
			name: "param exists",
			given: PathValues{
				{Name: "uid", Value: "101"},
				{Name: "fid", Value: "501"},
			},
			whenParamName:    "uid",
			whenDefaultValue: "999",
			expect:           "101",
		},
		{
			name: "param exists and is empty",
			given: PathValues{
				{Name: "uid", Value: ""},
				{Name: "fid", Value: "501"},
			},
			whenParamName:    "uid",
			whenDefaultValue: "999",
			expect:           "", // <-- this is different from QueryParamOr behaviour
		},
		{
			name: "param does not exists",
			given: PathValues{
				{Name: "uid", Value: "101"},
			},
			whenParamName:    "nope",
			whenDefaultValue: "999",
			expect:           "999",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			c := e.NewContext(req, nil)

			c.SetPathValues(tc.given)

			assert.EqualValues(t, tc.expect, c.ParamOr(tc.whenParamName, tc.whenDefaultValue))
		})
	}
}

func TestContextGetAndSetPathValuesMutability(t *testing.T) {
	t.Run("c.PathValues() does not return copy and modifying raw slice mutates value in context", func(t *testing.T) {
		e := New()
		e.contextPathParamAllocSize.Store(1)

		req := httptest.NewRequest(http.MethodGet, "/:foo", nil)
		c := e.NewContext(req, nil)

		params := PathValues{{Name: "foo", Value: "101"}}
		c.SetPathValues(params)

		// round-trip param values with modification
		paramVals := c.PathValues()
		assert.Equal(t, params, c.PathValues())

		// PathValues() does not return copy and modifying raw slice mutates value in context
		paramVals[0] = PathValue{Name: "xxx", Value: "yyy"}
		assert.Equal(t, PathValues{PathValue{Name: "xxx", Value: "yyy"}}, c.PathValues())
	})

	t.Run("calling SetPathValues with bigger size changes capacity in context", func(t *testing.T) {
		e := New()
		e.contextPathParamAllocSize.Store(1)

		req := httptest.NewRequest(http.MethodGet, "/:foo", nil)
		c := e.NewContext(req, nil)
		// increase path param capacity in context
		pathValues := PathValues{
			{Name: "aaa", Value: "bbb"},
			{Name: "ccc", Value: "ddd"},
		}
		c.SetPathValues(pathValues)
		assert.Equal(t, pathValues, c.PathValues())

		// shouldn't explode during Reset() afterwards!
		assert.NotPanics(t, func() {
			c.Reset(nil, nil)
		})
		assert.Equal(t, PathValues{}, c.PathValues())
		assert.Len(t, *c.pathValues, 0)
		assert.Equal(t, 2, cap(*c.pathValues))
	})

	t.Run("calling SetPathValues with smaller size slice does not change capacity in context", func(t *testing.T) {
		e := New()

		req := httptest.NewRequest(http.MethodGet, "/:foo", nil)
		c := e.NewContext(req, nil)
		c.pathValues = &PathValues{
			{Name: "aaa", Value: "bbb"},
			{Name: "ccc", Value: "ddd"},
		}

		pathValues := PathValues{
			{Name: "aaa", Value: "bbb"},
		}
		// given pathValues slice is smaller. this should not decrease c.pathValues capacity
		c.SetPathValues(pathValues)
		assert.Equal(t, pathValues, c.PathValues())

		// shouldn't explode during Reset() afterwards!
		assert.NotPanics(t, func() {
			c.Reset(nil, nil)
		})
		assert.Equal(t, PathValues{}, c.PathValues())
		assert.Len(t, *c.pathValues, 0)
		assert.Equal(t, 2, cap(*c.pathValues))
	})

}

// Issue #1655
func TestContext_SetParamNamesShouldNotModifyPathValuesCapacity(t *testing.T) {
	e := New()
	c := e.NewContext(nil, nil)

	assert.Equal(t, int32(0), e.contextPathParamAllocSize.Load())
	expectedTwoParams := PathValues{
		{Name: "1", Value: "one"},
		{Name: "2", Value: "two"},
	}
	c.SetPathValues(expectedTwoParams)
	assert.Equal(t, int32(0), e.contextPathParamAllocSize.Load())
	assert.Equal(t, expectedTwoParams, c.PathValues())

	expectedThreeParams := PathValues{
		{Name: "1", Value: "one"},
		{Name: "2", Value: "two"},
		{Name: "3", Value: "three"},
	}
	c.SetPathValues(expectedThreeParams)
	assert.Equal(t, int32(0), e.contextPathParamAllocSize.Load())
	assert.Equal(t, expectedThreeParams, c.PathValues())
}

func TestContextFormValue(t *testing.T) {
	f := make(url.Values)
	f.Set("name", "Jon Snow")
	f.Set("email", "jon@labstack.com")

	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Add(HeaderContentType, MIMEApplicationForm)
	c := e.NewContext(req, nil)

	// FormValue
	assert.Equal(t, "Jon Snow", c.FormValue("name"))
	assert.Equal(t, "jon@labstack.com", c.FormValue("email"))

	// FormValueOr
	assert.Equal(t, "Jon Snow", c.FormValueOr("name", "nope"))
	assert.Equal(t, "default", c.FormValueOr("missing", "default"))

	// FormValues
	values, err := c.FormValues()
	if assert.NoError(t, err) {
		assert.Equal(t, url.Values{
			"name":  []string{"Jon Snow"},
			"email": []string{"jon@labstack.com"},
		}, values)
	}

	// Multipart FormParams error
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Add(HeaderContentType, MIMEMultipartForm)
	c = e.NewContext(req, nil)
	values, err = c.FormValues()
	assert.Nil(t, values)
	assert.Error(t, err)
}

func TestContext_QueryParams(t *testing.T) {
	var testCases = []struct {
		expect   url.Values
		name     string
		givenURL string
	}{
		{
			name:     "multiple values in url",
			givenURL: "/?test=1&test=2&email=jon%40labstack.com",
			expect: url.Values{
				"test":  []string{"1", "2"},
				"email": []string{"jon@labstack.com"},
			},
		},
		{
			name:     "single value  in url",
			givenURL: "/?nope=1",
			expect: url.Values{
				"nope": []string{"1"},
			},
		},
		{
			name:     "no query params in url",
			givenURL: "/?",
			expect:   url.Values{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.givenURL, nil)
			e := New()
			c := e.NewContext(req, nil)

			assert.Equal(t, tc.expect, c.QueryParams())
		})
	}
}

func TestContext_QueryParam(t *testing.T) {
	var testCases = []struct {
		name          string
		givenURL      string
		whenParamName string
		expect        string
	}{
		{
			name:          "value exists in url",
			givenURL:      "/?test=1",
			whenParamName: "test",
			expect:        "1",
		},
		{
			name:          "multiple values exists in url",
			givenURL:      "/?test=9&test=8",
			whenParamName: "test",
			expect:        "9", // <-- first value in returned
		},
		{
			name:          "value does not exists in url",
			givenURL:      "/?nope=1",
			whenParamName: "test",
			expect:        "",
		},
		{
			name:          "value is empty in url",
			givenURL:      "/?test=",
			whenParamName: "test",
			expect:        "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.givenURL, nil)
			e := New()
			c := e.NewContext(req, nil)

			assert.Equal(t, tc.expect, c.QueryParam(tc.whenParamName))
		})
	}
}

func TestContext_QueryParamDefault(t *testing.T) {
	var testCases = []struct {
		name             string
		givenURL         string
		whenParamName    string
		whenDefaultValue string
		expect           string
	}{
		{
			name:             "value exists in url",
			givenURL:         "/?test=1",
			whenParamName:    "test",
			whenDefaultValue: "999",
			expect:           "1",
		},
		{
			name:             "value does not exists in url",
			givenURL:         "/?nope=1",
			whenParamName:    "test",
			whenDefaultValue: "999",
			expect:           "999",
		},
		{
			name:             "value is empty in url",
			givenURL:         "/?test=",
			whenParamName:    "test",
			whenDefaultValue: "999",
			expect:           "999",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.givenURL, nil)
			e := New()
			c := e.NewContext(req, nil)

			assert.Equal(t, tc.expect, c.QueryParamOr(tc.whenParamName, tc.whenDefaultValue))
		})
	}
}

func TestContextFormFile(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)
	mr := multipart.NewWriter(buf)
	w, err := mr.CreateFormFile("file", "test")
	if assert.NoError(t, err) {
		w.Write([]byte("test"))
	}
	mr.Close()
	req := httptest.NewRequest(http.MethodPost, "/", buf)
	req.Header.Set(HeaderContentType, mr.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	f, err := c.FormFile("file")
	if assert.NoError(t, err) {
		assert.Equal(t, "test", f.Filename)
	}
}

func TestContextMultipartForm(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	mw.WriteField("name", "Jon Snow")
	fileContent := "This is a test file"
	w, err := mw.CreateFormFile("file", "test.txt")
	if assert.NoError(t, err) {
		w.Write([]byte(fileContent))
	}
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/", buf)
	req.Header.Set(HeaderContentType, mw.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	f, err := c.MultipartForm()
	if assert.NoError(t, err) {
		assert.NotNil(t, f)

		files := f.File["file"]
		if assert.Len(t, files, 1) {
			file := files[0]
			assert.Equal(t, "test.txt", file.Filename)
			assert.Equal(t, int64(len(fileContent)), file.Size)
		}
	}
}

func TestContextRedirect(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	assert.Equal(t, nil, c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo"))
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "http://labstack.github.io/echo", rec.Header().Get(HeaderLocation))
	assert.Error(t, c.Redirect(310, "http://labstack.github.io/echo"))
}

func TestContextGet(t *testing.T) {
	var testCases = []struct {
		name    string
		given   any
		whenKey string
		expect  any
	}{
		{
			name:    "ok, value exist",
			given:   "Jon Snow",
			whenKey: "key",
			expect:  "Jon Snow",
		},
		{
			name:    "ok, value does not exist",
			given:   "Jon Snow",
			whenKey: "nope",
			expect:  nil,
		},
		{
			name:    "ok, value is nil value",
			given:   []byte(nil),
			whenKey: "key",
			expect:  []byte(nil),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var c = new(Context)
			c.Set("key", tc.given)

			v := c.Get(tc.whenKey)
			assert.Equal(t, tc.expect, v)
		})
	}
}

func BenchmarkContext_Store(b *testing.B) {
	e := &Echo{}

	c := &Context{
		echo: e,
	}

	for n := 0; n < b.N; n++ {
		c.Set("name", "Jon Snow")
		if c.Get("name") != "Jon Snow" {
			b.Fail()
		}
	}
}

type validator struct{}

func (*validator) Validate(i any) error {
	return nil
}

func TestContext_Validate(t *testing.T) {
	e := New()
	c := e.NewContext(nil, nil)

	assert.Error(t, c.Validate(struct{}{}))

	e.Validator = &validator{}
	assert.NoError(t, c.Validate(struct{}{}))
}

func TestContext_QueryString(t *testing.T) {
	e := New()

	queryString := "query=string&var=val"

	req := httptest.NewRequest(http.MethodGet, "/?"+queryString, nil)
	c := e.NewContext(req, nil)

	assert.Equal(t, queryString, c.QueryString())
}

func TestContext_Request(t *testing.T) {
	var c = new(Context)

	assert.Nil(t, c.Request())

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	c.SetRequest(req)

	assert.Equal(t, req, c.Request())
}

func TestContext_Scheme(t *testing.T) {
	tests := []struct {
		c *Context
		s string
	}{
		{
			&Context{
				request: &http.Request{
					TLS: &tls.ConnectionState{},
				},
			},
			"https",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedProto: []string{"https"}},
				},
			},
			"https",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedProtocol: []string{"http"}},
				},
			},
			"http",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedSsl: []string{"on"}},
				},
			},
			"https",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderXUrlScheme: []string{"https"}},
				},
			},
			"https",
		},
		{
			&Context{
				request: &http.Request{},
			},
			"http",
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.s, tt.c.Scheme())
	}
}

func TestContext_IsWebSocket(t *testing.T) {
	tests := []struct {
		c  *Context
		ws assert.BoolAssertionFunc
	}{
		{
			&Context{
				request: &http.Request{
					Header: http.Header{
						HeaderUpgrade:    []string{"websocket"},
						HeaderConnection: []string{"upgrade"},
					},
				},
			},
			assert.True,
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{
						HeaderUpgrade:    []string{"Websocket"},
						HeaderConnection: []string{"Upgrade"},
					},
				},
			},
			assert.True,
		},
		{
			&Context{
				request: &http.Request{},
			},
			assert.False,
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderUpgrade: []string{"other"}},
				},
			},
			assert.False,
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{
						HeaderUpgrade:    []string{"websocket"},
						HeaderConnection: []string{"close"},
					},
				},
			},
			assert.False,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %d", i+1), func(t *testing.T) {
			tt.ws(t, tt.c.IsWebSocket())
		})
	}
}

func TestContext_Bind(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	c := e.NewContext(req, nil)
	u := new(user)

	req.Header.Add(HeaderContentType, MIMEApplicationJSON)
	err := c.Bind(u)
	assert.NoError(t, err)
	assert.Equal(t, &user{ID: 1, Name: "Jon Snow"}, u)
}

func TestContext_RealIP(t *testing.T) {
	tests := []struct {
		c *Context
		s string
	}{
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1, 127.0.1.1, "}},
				},
			},
			"127.0.0.1",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1,127.0.1.1"}},
				},
			},
			"127.0.0.1",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1"}},
				},
			},
			"127.0.0.1",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedFor: []string{"[2001:db8:85a3:8d3:1319:8a2e:370:7348], 2001:db8::1, "}},
				},
			},
			"2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedFor: []string{"[2001:db8:85a3:8d3:1319:8a2e:370:7348],[2001:db8::1]"}},
				},
			},
			"2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedFor: []string{"2001:db8:85a3:8d3:1319:8a2e:370:7348"}},
				},
			},
			"2001:db8:85a3:8d3:1319:8a2e:370:7348",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{
						"X-Real-Ip": []string{"192.168.0.1"},
					},
				},
			},
			"192.168.0.1",
		},
		{
			&Context{
				request: &http.Request{
					Header: http.Header{
						"X-Real-Ip": []string{"[2001:db8::1]"},
					},
				},
			},
			"2001:db8::1",
		},

		{
			&Context{
				request: &http.Request{
					RemoteAddr: "89.89.89.89:1654",
				},
			},
			"89.89.89.89",
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.s, tt.c.RealIP())
	}
}

func TestContext_File(t *testing.T) {
	var testCases = []struct {
		whenFS           fs.FS
		name             string
		whenFile         string
		expectError      string
		expectStartsWith []byte
		expectStatus     int
	}{
		{
			name:             "ok, from default file system",
			whenFile:         "_fixture/images/walle.png",
			whenFS:           nil,
			expectStatus:     http.StatusOK,
			expectStartsWith: []byte{0x89, 0x50, 0x4e},
		},
		{
			name:             "ok, from custom file system",
			whenFile:         "walle.png",
			whenFS:           os.DirFS("_fixture/images"),
			expectStatus:     http.StatusOK,
			expectStartsWith: []byte{0x89, 0x50, 0x4e},
		},
		{
			name:             "nok, not existent file",
			whenFile:         "not.png",
			whenFS:           os.DirFS("_fixture/images"),
			expectStatus:     http.StatusOK,
			expectStartsWith: nil,
			expectError:      "Not Found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			if tc.whenFS != nil {
				e.Filesystem = tc.whenFS
			}

			handler := func(ec *Context) error {
				return ec.File(tc.whenFile)
			}

			req := httptest.NewRequest(http.MethodGet, "/match.png", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler(c)

			assert.Equal(t, tc.expectStatus, rec.Code)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}

			body := rec.Body.Bytes()
			if len(body) > len(tc.expectStartsWith) {
				body = body[:len(tc.expectStartsWith)]
			}
			assert.Equal(t, tc.expectStartsWith, body)
		})
	}
}

func TestContext_FileFS(t *testing.T) {
	var testCases = []struct {
		whenFS           fs.FS
		name             string
		whenFile         string
		expectError      string
		expectStartsWith []byte
		expectStatus     int
	}{
		{
			name:             "ok",
			whenFile:         "walle.png",
			whenFS:           os.DirFS("_fixture/images"),
			expectStatus:     http.StatusOK,
			expectStartsWith: []byte{0x89, 0x50, 0x4e},
		},
		{
			name:             "nok, not existent file",
			whenFile:         "not.png",
			whenFS:           os.DirFS("_fixture/images"),
			expectStatus:     http.StatusOK,
			expectStartsWith: nil,
			expectError:      "Not Found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			handler := func(ec *Context) error {
				return ec.FileFS(tc.whenFile, tc.whenFS)
			}

			req := httptest.NewRequest(http.MethodGet, "/match.png", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler(c)

			assert.Equal(t, tc.expectStatus, rec.Code)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}

			body := rec.Body.Bytes()
			if len(body) > len(tc.expectStartsWith) {
				body = body[:len(tc.expectStartsWith)]
			}
			assert.Equal(t, tc.expectStartsWith, body)
		})
	}
}

func TestLogger(t *testing.T) {
	e := New()
	c := e.NewContext(nil, nil)

	log1 := c.Logger()
	assert.NotNil(t, log1)
	assert.Equal(t, e.Logger, log1)

	customLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	c.SetLogger(customLogger)
	assert.Equal(t, customLogger, c.Logger())

	// Resetting the context returns the initial Echo logger
	c.Reset(nil, nil)
	assert.Equal(t, e.Logger, c.Logger())
}

func TestRouteInfo(t *testing.T) {
	e := New()
	c := e.NewContext(nil, nil)

	orgRI := RouteInfo{
		Name:       "root",
		Method:     http.MethodGet,
		Path:       "/*",
		Parameters: []string{"*"},
	}
	c.route = &orgRI
	ri := c.RouteInfo()
	assert.Equal(t, orgRI, ri)

	// Test mutability when middlewares start to change things

	// RouteInfo inside context will not be affected when returned instance is changed
	expect := orgRI.Clone()
	ri.Path = "changed"
	ri.Parameters[0] = "changed"
	assert.Equal(t, expect, c.RouteInfo())

	// RouteInfo inside context will not be affected when returned instance is changed
	expect = c.RouteInfo()
	orgRI.Name = "changed"
	assert.NotEqual(t, expect, c.RouteInfo())
}
