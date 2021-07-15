package echo

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

var testUser = user{1, "Jon Snow"}

func BenchmarkAllocJSONP(b *testing.B) {
	e := New()
	e.Logger = &noOpLogger{}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*DefaultContext)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.JSONP(http.StatusOK, "callback", testUser)
	}
}

func BenchmarkAllocJSON(b *testing.B) {
	e := New()
	e.Logger = &noOpLogger{}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*DefaultContext)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.JSON(http.StatusOK, testUser)
	}
}

func BenchmarkAllocXML(b *testing.B) {
	e := New()
	e.Logger = &noOpLogger{}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*DefaultContext)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.XML(http.StatusOK, testUser)
	}
}

func BenchmarkRealIPForHeaderXForwardFor(b *testing.B) {
	c := DefaultContext{request: &http.Request{
		Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1, 127.0.1.1, "}},
	}}
	for i := 0; i < b.N; i++ {
		c.RealIP()
	}
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type responseWriterErr struct {
}

func (responseWriterErr) Header() http.Header {
	return http.Header{}
}

func (responseWriterErr) Write([]byte) (int, error) {
	return 0, errors.New("err")
}

func (responseWriterErr) WriteHeader(statusCode int) {

}

func TestContext(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*DefaultContext)

	// Echo
	assert.Equal(t, e, c.Echo())

	// Request
	assert.NotNil(t, c.Request())

	// Response
	assert.NotNil(t, c.Response())

	//--------
	// Render
	//--------

	tmpl := &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	c.echo.Renderer = tmpl
	err := c.Render(http.StatusOK, "hello", "Jon Snow")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello, Jon Snow!", rec.Body.String())
	}

	c.echo.Renderer = nil
	err = c.Render(http.StatusOK, "hello", "Jon Snow")
	assert.Error(t, err)

	// JSON
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.JSON(http.StatusOK, user{1, "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJSONCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSON+"\n", rec.Body.String())
	}

	// JSON with "?pretty"
	req = httptest.NewRequest(http.MethodGet, "/?pretty", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.JSON(http.StatusOK, user{1, "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJSONCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSONPretty+"\n", rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil) // reset

	// JSONPretty
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.JSONPretty(http.StatusOK, user{1, "Jon Snow"}, "  ")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJSONCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSONPretty+"\n", rec.Body.String())
	}

	// JSON (error)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.JSON(http.StatusOK, make(chan bool))
	assert.Error(t, err)

	// JSONP
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	callback := "callback"
	err = c.JSONP(http.StatusOK, callback, user{1, "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJavaScriptCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, callback+"("+userJSON+"\n);", rec.Body.String())
	}

	// XML
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.XML(http.StatusOK, user{1, "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+userXML, rec.Body.String())
	}

	// XML with "?pretty"
	req = httptest.NewRequest(http.MethodGet, "/?pretty", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.XML(http.StatusOK, user{1, "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+userXMLPretty, rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)

	// XML (error)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.XML(http.StatusOK, make(chan bool))
	assert.Error(t, err)

	// XML response write error
	c = e.NewContext(req, rec).(*DefaultContext)
	c.response.Writer = responseWriterErr{}
	err = c.XML(0, 0)
	assert.Error(t, err)

	// XMLPretty
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.XMLPretty(http.StatusOK, user{1, "Jon Snow"}, "  ")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+userXMLPretty, rec.Body.String())
	}

	t.Run("empty indent", func(t *testing.T) {
		var (
			u           = user{1, "Jon Snow"}
			buf         = new(bytes.Buffer)
			emptyIndent = ""
		)

		t.Run("json", func(t *testing.T) {
			buf.Reset()

			// New JSONBlob with empty indent
			rec = httptest.NewRecorder()
			c = e.NewContext(req, rec).(*DefaultContext)
			enc := json.NewEncoder(buf)
			enc.SetIndent(emptyIndent, emptyIndent)
			err = enc.Encode(u)
			err = c.json(http.StatusOK, user{1, "Jon Snow"}, emptyIndent)
			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, MIMEApplicationJSONCharsetUTF8, rec.Header().Get(HeaderContentType))
				assert.Equal(t, buf.String(), rec.Body.String())
			}
		})

		t.Run("xml", func(t *testing.T) {
			buf.Reset()

			// New XMLBlob with empty indent
			rec = httptest.NewRecorder()
			c = e.NewContext(req, rec).(*DefaultContext)
			enc := xml.NewEncoder(buf)
			enc.Indent(emptyIndent, emptyIndent)
			err = enc.Encode(u)
			err = c.xml(http.StatusOK, user{1, "Jon Snow"}, emptyIndent)
			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
				assert.Equal(t, xml.Header+buf.String(), rec.Body.String())
			}
		})
	})

	// Legacy JSONBlob
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	data, err := json.Marshal(user{1, "Jon Snow"})
	assert.NoError(t, err)
	err = c.JSONBlob(http.StatusOK, data)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJSONCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSON, rec.Body.String())
	}

	// Legacy JSONPBlob
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	callback = "callback"
	data, err = json.Marshal(user{1, "Jon Snow"})
	assert.NoError(t, err)
	err = c.JSONPBlob(http.StatusOK, callback, data)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJavaScriptCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, callback+"("+userJSON+");", rec.Body.String())
	}

	// Legacy XMLBlob
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	data, err = xml.Marshal(user{1, "Jon Snow"})
	assert.NoError(t, err)
	err = c.XMLBlob(http.StatusOK, data)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+userXML, rec.Body.String())
	}

	// String
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.String(http.StatusOK, "Hello, World!")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMETextPlainCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}

	// HTML
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.HTML(http.StatusOK, "Hello, <strong>World!</strong>")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMETextHTMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, "Hello, <strong>World!</strong>", rec.Body.String())
	}

	// Stream
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	r := strings.NewReader("response from a stream")
	err = c.Stream(http.StatusOK, "application/octet-stream", r)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/octet-stream", rec.Header().Get(HeaderContentType))
		assert.Equal(t, "response from a stream", rec.Body.String())
	}

	// Attachment
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.Attachment("_fixture/images/walle.png", "walle.png")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "attachment; filename=\"walle.png\"", rec.Header().Get(HeaderContentDisposition))
		assert.Equal(t, 219885, rec.Body.Len())
	}

	// Inline
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	err = c.Inline("_fixture/images/walle.png", "walle.png")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "inline; filename=\"walle.png\"", rec.Header().Get(HeaderContentDisposition))
		assert.Equal(t, 219885, rec.Body.Len())
	}

	// NoContent
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*DefaultContext)
	c.NoContent(http.StatusOK)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Reset
	c.pathParams = &PathParams{
		{Name: "foo", Value: "bar"},
	}
	c.Set("foe", "ban")
	c.query = url.Values(map[string][]string{"fon": {"baz"}})
	c.Reset(req, httptest.NewRecorder())
	assert.Equal(t, 0, len(c.PathParams()))
	assert.Equal(t, 0, len(c.store))
	assert.Equal(t, nil, c.RouteInfo())
	assert.Equal(t, 0, len(c.QueryParams()))
}

func TestContext_JSON_CommitsCustomResponseCode(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*DefaultContext)
	err := c.JSON(http.StatusCreated, user{1, "Jon Snow"})

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, MIMEApplicationJSONCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSON+"\n", rec.Body.String())
	}
}

func TestContext_JSON_DoesntCommitResponseCodePrematurely(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*DefaultContext)
	err := c.JSON(http.StatusCreated, map[string]float64{"a": math.NaN()})

	if assert.Error(t, err) {
		assert.False(t, c.response.Committed)
	}
}

func TestContextCookie(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	theme := "theme=light"
	user := "user=Jon Snow"
	req.Header.Add(HeaderCookie, theme)
	req.Header.Add(HeaderCookie, user)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*DefaultContext)

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

func TestContextPathParam(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, nil).(*DefaultContext)

	params := &PathParams{
		{Name: "uid", Value: "101"},
		{Name: "fid", Value: "501"},
	}
	// ParamNames
	c.pathParams = params
	assert.EqualValues(t, *params, c.PathParams())

	// Param
	assert.Equal(t, "501", c.PathParam("fid"))
	assert.Equal(t, "", c.PathParam("undefined"))
}

func TestContextGetAndSetParam(t *testing.T) {
	e := New()
	r := e.Router()
	_, err := r.Add(Route{
		Method:      http.MethodGet,
		Path:        "/:foo",
		Name:        "",
		Handler:     func(Context) error { return nil },
		Middlewares: nil,
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/:foo", nil)
	c := e.NewContext(req, nil)

	params := &PathParams{{Name: "foo", Value: "101"}}
	// ParamNames
	c.(*DefaultContext).pathParams = params

	// round-trip param values with modification
	paramVals := c.PathParams()
	assert.Equal(t, *params, c.PathParams())

	paramVals[0] = PathParam{Name: "xxx", Value: "yyy"} // PathParams() returns copy and modifying it does nothing to context
	assert.Equal(t, PathParams{{Name: "foo", Value: "101"}}, c.PathParams())

	pathParams := PathParams{
		{Name: "aaa", Value: "bbb"},
		{Name: "ccc", Value: "ddd"},
	}
	c.SetPathParams(pathParams)
	assert.Equal(t, pathParams, c.PathParams())

	// shouldn't explode during Reset() afterwards!
	assert.NotPanics(t, func() {
		c.(ServableContext).Reset(nil, nil)
	})
	assert.Equal(t, PathParams{}, c.PathParams())
	assert.Len(t, *c.(*DefaultContext).pathParams, 0)
	assert.Equal(t, cap(*c.(*DefaultContext).pathParams), 1)
}

// Issue #1655
func TestContext_SetParamNamesShouldNotModifyPathParams(t *testing.T) {
	e := New()
	c := e.NewContext(nil, nil).(*DefaultContext)

	assert.Equal(t, 0, e.contextPathParamAllocSize)
	expectedTwoParams := &PathParams{
		{Name: "1", Value: "one"},
		{Name: "2", Value: "two"},
	}
	c.SetRawPathParams(expectedTwoParams)
	assert.Equal(t, 0, e.contextPathParamAllocSize)
	assert.Equal(t, *expectedTwoParams, c.PathParams())

	expectedThreeParams := PathParams{
		{Name: "1", Value: "one"},
		{Name: "2", Value: "two"},
		{Name: "3", Value: "three"},
	}
	c.SetPathParams(expectedThreeParams)
	assert.Equal(t, 0, e.contextPathParamAllocSize)
	assert.Equal(t, expectedThreeParams, c.PathParams())
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

	// FormValueDefault
	assert.Equal(t, "Jon Snow", c.FormValueDefault("name", "nope"))
	assert.Equal(t, "default", c.FormValueDefault("missing", "default"))

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

func TestContextQueryParam(t *testing.T) {
	q := make(url.Values)
	q.Set("name", "Jon Snow")
	q.Set("email", "jon@labstack.com")
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	e := New()
	c := e.NewContext(req, nil)

	// QueryParam
	assert.Equal(t, "Jon Snow", c.QueryParam("name"))
	assert.Equal(t, "jon@labstack.com", c.QueryParam("email"))

	// QueryParamDefault
	assert.Equal(t, "Jon Snow", c.QueryParamDefault("name", "nope"))
	assert.Equal(t, "default", c.QueryParamDefault("missing", "default"))

	// QueryParams
	assert.Equal(t, url.Values{
		"name":  []string{"Jon Snow"},
		"email": []string{"jon@labstack.com"},
	}, c.QueryParams())
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
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/", buf)
	req.Header.Set(HeaderContentType, mw.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	f, err := c.MultipartForm()
	if assert.NoError(t, err) {
		assert.NotNil(t, f)
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

func TestContextStore(t *testing.T) {
	var c Context = new(DefaultContext)
	c.Set("name", "Jon Snow")
	assert.Equal(t, "Jon Snow", c.Get("name"))
}

func BenchmarkContext_Store(b *testing.B) {
	e := &Echo{}

	c := &DefaultContext{
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

func (*validator) Validate(i interface{}) error {
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
	var c Context = new(DefaultContext)

	assert.Nil(t, c.Request())

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	c.SetRequest(req)

	assert.Equal(t, req, c.Request())
}

func TestContext_Scheme(t *testing.T) {
	tests := []struct {
		c Context
		s string
	}{
		{
			&DefaultContext{
				request: &http.Request{
					TLS: &tls.ConnectionState{},
				},
			},
			"https",
		},
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedProto: []string{"https"}},
				},
			},
			"https",
		},
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedProtocol: []string{"http"}},
				},
			},
			"http",
		},
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedSsl: []string{"on"}},
				},
			},
			"https",
		},
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{HeaderXUrlScheme: []string{"https"}},
				},
			},
			"https",
		},
		{
			&DefaultContext{
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
		c  Context
		ws assert.BoolAssertionFunc
	}{
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{HeaderUpgrade: []string{"websocket"}},
				},
			},
			assert.True,
		},
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{HeaderUpgrade: []string{"Websocket"}},
				},
			},
			assert.True,
		},
		{
			&DefaultContext{
				request: &http.Request{},
			},
			assert.False,
		},
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{HeaderUpgrade: []string{"other"}},
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
	assert.Equal(t, &user{1, "Jon Snow"}, u)
}

func TestContext_RealIP(t *testing.T) {
	tests := []struct {
		c Context
		s string
	}{
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1, 127.0.1.1, "}},
				},
			},
			"127.0.0.1",
		},
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1,127.0.1.1"}},
				},
			},
			"127.0.0.1",
		},
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{HeaderXForwardedFor: []string{"127.0.0.1"}},
				},
			},
			"127.0.0.1",
		},
		{
			&DefaultContext{
				request: &http.Request{
					Header: http.Header{
						"X-Real-Ip": []string{"192.168.0.1"},
					},
				},
			},
			"192.168.0.1",
		},
		{
			&DefaultContext{
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
		name             string
		whenFile         string
		whenFS           fs.FS
		expectStatus     int
		expectStartsWith []byte
		expectError      string
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
			expectError:      "code=404, message=Not Found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			if tc.whenFS != nil {
				e.Filesystem = tc.whenFS
			}

			handler := func(ec Context) error {
				return ec.(*DefaultContext).File(tc.whenFile)
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
		name             string
		whenFile         string
		whenFS           fs.FS
		expectStatus     int
		expectStartsWith []byte
		expectError      string
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
			expectError:      "code=404, message=Not Found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()

			handler := func(ec Context) error {
				return ec.(*DefaultContext).FileFS(tc.whenFile, tc.whenFS)
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
