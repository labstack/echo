package echo

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"strings"

	"golang.org/x/net/context"

	"net/url"

	"encoding/xml"

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
	userJSONIndent := "{\n_?\"id\": \"1\",\n_?\"name\": \"Joe\"\n_}"
	userXML := `<user><id>1</id><name>Joe</name></user>`
	userXMLIndent := "_<user>\n_?<id>1</id>\n_?<name>Joe</name>\n_</user>"

	var nonMarshallableChannel chan bool

	e := New()
	req, _ := http.NewRequest(POST, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := NewContext(req, NewResponse(rec, e), e)

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
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.JSON(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, ApplicationJSONCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, userJSON, rec.Body.String())
	}

	// JSON (error)
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	val := make(chan bool)
	err = c.JSON(http.StatusOK, val)
	assert.Error(t, err)

	// JSONIndent
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.JSONIndent(http.StatusOK, user{"1", "Joe"}, "_", "?")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, ApplicationJSONCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, userJSONIndent, rec.Body.String())
	}

	// JSONIndent (error)
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.JSONIndent(http.StatusOK, nonMarshallableChannel, "_", "?")
	assert.Error(t, err)

	// JSONP
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	callback := "callback"
	err = c.JSONP(http.StatusOK, callback, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, ApplicationJavaScriptCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, callback+"("+userJSON+");", rec.Body.String())
	}

	// XML
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.XML(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, ApplicationXMLCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, xml.Header+userXML, rec.Body.String())
	}

	// XML (error)
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.XML(http.StatusOK, nonMarshallableChannel)
	assert.Error(t, err)

	// XMLIndent
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.XMLIndent(http.StatusOK, user{"1", "Joe"}, "_", "?")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, ApplicationXMLCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, xml.Header+userXMLIndent, rec.Body.String())
	}

	// XMLIndent (error)
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.XMLIndent(http.StatusOK, nonMarshallableChannel, "_", "?")
	assert.Error(t, err)

	// String
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.String(http.StatusOK, "Hello, World!")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, TextPlainCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}

	// HTML
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.HTML(http.StatusOK, "Hello, <strong>World!</strong>")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, TextHTMLCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, "Hello, <strong>World!</strong>", rec.Body.String())
	}

	// File
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.File("_fixture/images/walle.png", "", false)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 219885, rec.Body.Len())
	}

	// File as attachment
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	err = c.File("_fixture/images/walle.png", "WALLE.PNG", true)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, rec.Header().Get(ContentDisposition), "attachment; filename=WALLE.PNG")
		assert.Equal(t, 219885, rec.Body.Len())
	}

	// NoContent
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	c.NoContent(http.StatusOK)
	assert.Equal(t, http.StatusOK, c.response.status)

	// Redirect
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	assert.Equal(t, nil, c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo"))

	// Error
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec, e), e)
	c.Error(errors.New("error"))
	assert.Equal(t, http.StatusInternalServerError, c.response.status)

	// reset
	c.reset(req, NewResponse(httptest.NewRecorder(), e), e)
}

func TestContextPath(t *testing.T) {
	e := New()
	r := e.Router()

	r.Add(GET, "/users/:id", nil, e)
	c := NewContext(nil, nil, e)
	r.Find(GET, "/users/1", c)
	assert.Equal(t, c.Path(), "/users/:id")

	r.Add(GET, "/users/:uid/files/:fid", nil, e)
	c = NewContext(nil, nil, e)
	r.Find(GET, "/users/1/files/1", c)
	assert.Equal(t, c.Path(), "/users/:uid/files/:fid")
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

func TestContextNetContext(t *testing.T) {
	c := new(Context)
	c.Context = context.WithValue(nil, "key", "val")
	assert.Equal(t, "val", c.Value("key"))
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
