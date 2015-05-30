package echo

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"strings"

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
	usr := `{"id":"1","name":"Joe"}`
	req, _ := http.NewRequest(POST, "/", strings.NewReader(usr))
	rec := httptest.NewRecorder()
	c := NewContext(req, NewResponse(rec), New())

	// Request
	assert.NotNil(t, c.Request())

	// Response
	assert.NotNil(t, c.Response())

	// Socket
	assert.Nil(t, c.Socket())

	//-------
	// Param
	//-------

	// By id
	c.pnames = []string{"id"}
	c.pvalues = []string{"1"}
	assert.Equal(t, "1", c.P(0))

	// By name
	assert.Equal(t, "1", c.Param("id"))

	// Store
	c.Set("user", "Joe")
	assert.Equal(t, "Joe", c.Get("user"))

	//------
	// Bind
	//------

	// JSON
	testBind(t, c, ApplicationJSON)

	// TODO: Form
	c.request.Header.Set(ContentType, ApplicationForm)
	u := new(user)
	err := c.Bind(u)
	assert.NoError(t, err)

	// Unsupported
	c.request.Header.Set(ContentType, "")
	u = new(user)
	err = c.Bind(u)
	assert.Error(t, err)

	//--------
	// Render
	//--------

	tpl := &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	c.echo.SetRenderer(tpl)
	err = c.Render(http.StatusOK, "hello", "Joe")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello, Joe!", rec.Body.String())
	}

	c.echo.renderer = nil
	err = c.Render(http.StatusOK, "hello", "Joe")
	assert.Error(t, err)

	// JSON
	req.Header.Set(Accept, ApplicationJSON)
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	err = c.JSON(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, ApplicationJSON, rec.Header().Get(ContentType))
		assert.Equal(t, usr, strings.TrimSpace(rec.Body.String()))
	}

	// String
	req.Header.Set(Accept, TextPlain)
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	err = c.String(http.StatusOK, "Hello, World!")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, TextPlain, rec.Header().Get(ContentType))
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}

	// HTML
	req.Header.Set(Accept, TextHTML)
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	err = c.HTML(http.StatusOK, "Hello, <strong>World!</strong>")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, TextHTML, rec.Header().Get(ContentType))
		assert.Equal(t, "Hello, <strong>World!</strong>", rec.Body.String())
	}

	// NoContent
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	c.NoContent(http.StatusOK)
	assert.Equal(t, http.StatusOK, c.response.status)

	// Redirect
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo")

	// Error
	rec = httptest.NewRecorder()
	c = NewContext(req, NewResponse(rec), New())
	c.Error(errors.New("error"))
	assert.Equal(t, http.StatusInternalServerError, c.response.status)

	// reset
	c.reset(req, NewResponse(httptest.NewRecorder()), New())
}

func testBind(t *testing.T, c *Context, ct string) {
	c.request.Header.Set(ContentType, ct)
	u := new(user)
	err := c.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, "1", u.ID)
		assert.Equal(t, "Joe", u.Name)
	}
}
