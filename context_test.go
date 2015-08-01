package echo

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"strings"

	"encoding/xml"
	"net/url"

	"github.com/stretchr/testify/assert"
)

type (
	Template struct {
		templates *template.Template
	}
)

func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func TestContext(t *testing.T) {
	userJSON := `{"id":"1","name":"Joe"}`
	userXML := `<user><id>1</id><name>Joe</name></user>`

	req, _ := http.NewRequest(POST, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := NewContext(req, NewResponse(rec), New())

	// Request
	assert.NotNil(t, c.Request())

	// Response
	assert.NotNil(t, c.Response())

	// Socket
	assert.Nil(t, c.Socket())

	// Param by id
	c.pnames = []string{"id"}
	c.pvalues = []string{"1"}
	assert.Equal(t, "1", c.P(0))

	// Param by name
	assert.Equal(t, "1", c.Param("id"))

	// Store
	c.Set("user", "Joe")
	assert.Equal(t, "Joe", c.Get("user"))

	//------
	// Bind
	//------

	// JSON
	testBind(t, c, "application/json")

	// XML
	c.request, _ = http.NewRequest(POST, "/", strings.NewReader(userXML))
	testBind(t, c, ApplicationXML)

	// Unsupported
	testBind(t, c, "")

	//--------
	// Render
	//--------

	tpl := &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	c.echo.SetRenderer(tpl)
	err := c.Render(http.StatusOK, "hello", "Joe")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello, Joe!", rec.Body.String())
	}

	c.echo.renderer = nil
	err = c.Render(http.StatusOK, "hello", "Joe")
	assert.Error(t, err)

	// JSON
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	err = c.JSON(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, ApplicationJSONCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, userJSON+"\n", rec.Body.String())
	}

	// JSONP
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	callback := "callback"
	err = c.JSONP(http.StatusOK, callback, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, ApplicationJavaScriptCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, callback+"("+userJSON+"\n);", rec.Body.String())
	}

	// XML
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	err = c.XML(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, ApplicationXMLCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, xml.Header, xml.Header, rec.Body.String())
	}

	// String
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	err = c.String(http.StatusOK, "Hello, World!")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, TextPlain, rec.Header().Get(ContentType))
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}

	// HTML
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	err = c.HTML(http.StatusOK, "Hello, <strong>World!</strong>")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, TextHTMLCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, "Hello, <strong>World!</strong>", rec.Body.String())
	}

	// File
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	err = c.File("test/fixture/walle.png")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 219885, rec.Body.Len())
	}

	// NoContent
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	c.NoContent(http.StatusOK)
	assert.Equal(t, http.StatusOK, c.response.status)

	// Redirect
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	assert.Equal(t, nil, c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo"))

	// Error
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	c.Error(errors.New("error"))
	assert.Equal(t, http.StatusInternalServerError, c.response.status)

	// reset
	c.reset(req, NewResponse(httptest.NewRecorder()), New())
}

func TestContextQuery(t *testing.T) {
	q := make(url.Values)
	q.Set("name", "joe")
	q.Set("email", "joe@labstack.com")

	req, err := http.NewRequest(GET, "/", nil)
	assert.NoError(t, err)
	req.URL.RawQuery = q.Encode()

	c := NewContext(req, nil, New())
	assert.Equal(t, "joe", c.Query("name"))
	assert.Equal(t, "joe@labstack.com", c.Query("email"))

}

func TestContextForm(t *testing.T) {
	f := make(url.Values)
	f.Set("name", "joe")
	f.Set("email", "joe@labstack.com")

	req, err := http.NewRequest(POST, "/", strings.NewReader(f.Encode()))
	assert.NoError(t, err)
	req.Header.Add(ContentType, ApplicationForm)

	c := NewContext(req, nil, New())
	assert.Equal(t, "joe", c.Form("name"))
	assert.Equal(t, "joe@labstack.com", c.Form("email"))
}

func testBind(t *testing.T, c *Context, ct string) {
	c.request.Header.Set(ContentType, ct)
	u := new(user)
	err := c.Bind(u)
	if ct == "" {
		assert.Error(t, UnsupportedMediaType)
	} else if assert.NoError(t, err) {
		assert.Equal(t, "1", u.ID)
		assert.Equal(t, "Joe", u.Name)
	}
}
