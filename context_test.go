package echo

import (
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"text/template"

	"strings"

	"net/url"

	"encoding/xml"

	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

type (
	Template struct {
		templates *template.Template
	}
)

func (t *Template) Render(w io.Writer, name string, data interface{}, c Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func TestContext(t *testing.T) {
	userJSON := `{"id":"1","name":"Joe"}`
	userXML := `<user><id>1</id><name>Joe</name></user>`
	invalidContent := "invalid content"

	e := New()
	req := test.NewRequest(POST, "/", strings.NewReader(userJSON))
	rec := test.NewResponseRecorder()
	c := NewContext(req, rec, e)

	// Request
	assert.NotNil(t, c.Request())

	// Response
	assert.NotNil(t, c.Response())

	// ParamNames
	c.Object().pnames = []string{"uid", "fid"}
	assert.EqualValues(t, []string{"uid", "fid"}, c.ParamNames())

	// Param by id
	c.Object().pnames = []string{"id"}
	c.Object().pvalues = []string{"1"}
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
	testBindOk(t, c, ApplicationJSON)
	c.Object().request = test.NewRequest(POST, "/", strings.NewReader(invalidContent))
	testBindError(t, c, ApplicationJSON)

	// XML
	c.Object().request = test.NewRequest(POST, "/", strings.NewReader(userXML))
	testBindOk(t, c, ApplicationXML)
	c.Object().request = test.NewRequest(POST, "/", strings.NewReader(invalidContent))
	testBindError(t, c, ApplicationXML)

	// Unsupported
	testBindError(t, c, "")

	//--------
	// Render
	//--------

	tpl := &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	c.Object().echo.SetRenderer(tpl)
	err := c.Render(http.StatusOK, "hello", "Joe")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, "Hello, Joe!", rec.Body.String())
	}

	c.Object().echo.renderer = nil
	err = c.Render(http.StatusOK, "hello", "Joe")
	assert.Error(t, err)

	// JSON
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	err = c.JSON(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, ApplicationJSONCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, userJSON, rec.Body.String())
	}

	// JSON (error)
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	err = c.JSON(http.StatusOK, make(chan bool))
	assert.Error(t, err)

	// JSONP
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	callback := "callback"
	err = c.JSONP(http.StatusOK, callback, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, ApplicationJavaScriptCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, callback+"("+userJSON+");", rec.Body.String())
	}

	// XML
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	err = c.XML(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, ApplicationXMLCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, xml.Header+userXML, rec.Body.String())
	}

	// XML (error)
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	err = c.XML(http.StatusOK, make(chan bool))
	assert.Error(t, err)

	// String
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	err = c.String(http.StatusOK, "Hello, World!")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, TextPlainCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}

	// HTML
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	err = c.HTML(http.StatusOK, "Hello, <strong>World!</strong>")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, TextHTMLCharsetUTF8, rec.Header().Get(ContentType))
		assert.Equal(t, "Hello, <strong>World!</strong>", rec.Body.String())
	}

	// Attachment
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	file, err := os.Open("_fixture/images/walle.png")
	if assert.NoError(t, err) {
		err = c.Attachment(file, "walle.png")
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Status())
			assert.Equal(t, "attachment; filename=walle.png", rec.Header().Get(ContentDisposition))
			assert.Equal(t, 219885, rec.Body.Len())
		}
	}

	// NoContent
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	c.NoContent(http.StatusOK)
	assert.Equal(t, http.StatusOK, rec.Status())

	// Redirect
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	assert.Equal(t, nil, c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo"))
	assert.Equal(t, "http://labstack.github.io/echo", rec.Header().Get(Location))
	assert.Equal(t, http.StatusMovedPermanently, rec.Status())

	// Error
	rec = test.NewResponseRecorder()
	c = NewContext(req, rec, e)
	c.Error(errors.New("error"))
	assert.Equal(t, http.StatusInternalServerError, rec.Status())

	// reset
	c.Object().reset(req, test.NewResponseRecorder())
}

func TestContextPath(t *testing.T) {
	e := New()
	r := e.Router()

	r.Add(GET, "/users/:id", nil, e)
	c := NewContext(nil, nil, e)
	r.Find(GET, "/users/1", c)
	assert.Equal(t, "/users/:id", c.Path())

	r.Add(GET, "/users/:uid/files/:fid", nil, e)
	c = NewContext(nil, nil, e)
	r.Find(GET, "/users/1/files/1", c)
	assert.Equal(t, "/users/:uid/files/:fid", c.Path())
}

func TestContextQuery(t *testing.T) {
	q := make(url.Values)
	q.Set("name", "joe")
	q.Set("email", "joe@labstack.com")
	req := test.NewRequest(GET, "/?"+q.Encode(), nil)
	c := NewContext(req, nil, New())
	assert.Equal(t, "joe", c.Query("name"))
	assert.Equal(t, "joe@labstack.com", c.Query("email"))
}

func TestContextForm(t *testing.T) {
	f := make(url.Values)
	f.Set("name", "joe")
	f.Set("email", "joe@labstack.com")

	req := test.NewRequest(POST, "/", strings.NewReader(f.Encode()))
	req.Header().Add(ContentType, ApplicationForm)

	c := NewContext(req, nil, New())
	assert.Equal(t, "joe", c.Form("name"))
	assert.Equal(t, "joe@labstack.com", c.Form("email"))
}

func TestContextNetContext(t *testing.T) {
	// c := new(context)
	// c.Context = xcontext.WithValue(nil, "key", "val")
	// assert.Equal(t, "val", c.Value("key"))
}

func testBindOk(t *testing.T, c Context, ct string) {
	c.Request().Header().Set(ContentType, ct)
	u := new(user)
	err := c.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, "1", u.ID)
		assert.Equal(t, "Joe", u.Name)
	}
}

func testBindError(t *testing.T, c Context, ct string) {
	c.Request().Header().Set(ContentType, ct)
	u := new(user)
	err := c.Bind(u)

	switch ct {
	case ApplicationJSON, ApplicationXML:
		if assert.IsType(t, new(HTTPError), err) {
			assert.Equal(t, http.StatusBadRequest, err.(*HTTPError).Code)
		}
	default:
		if assert.IsType(t, new(HTTPError), err) {
			assert.Equal(t, ErrUnsupportedMediaType, err)
		}

	}
}
