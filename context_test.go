package echo

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
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
	b, _ := json.Marshal(u1)
	r, _ := http.NewRequest(POST, "/users/1", bytes.NewReader(b))
	c := NewContext(r, NewResponse(httptest.NewRecorder()), New())

	//------
	// Bind
	//------

	// JSON
	r.Header.Set(ContentType, ApplicationJSON)
	u2 := new(user)
	if he := c.Bind(u2); he != nil {
		t.Errorf("bind %#v", he)
	}
	verifyUser(u2, t)

	// FORM
	r.Header.Set(ContentType, ApplicationForm)
	u2 = new(user)
	if he := c.Bind(u2); he != nil {
		t.Errorf("bind %#v", he)
	}
	// TODO: add verification

	// Unsupported
	r.Header.Set(ContentType, "")
	u2 = new(user)
	if he := c.Bind(u2); he == nil {
		t.Errorf("bind %#v", he)
	}
	// TODO: add verification

	//-------
	// Param
	//-------

	// By id
	c.pnames = []string{"id"}
	c.pvalues = []string{"1"}
	if c.P(0) != "1" {
		t.Error("param id should be 1")
	}

	// By name
	if c.Param("id") != "1" {
		t.Error("param id should be 1")
	}

	// Store
	c.Set("user", u1.Name)
	n := c.Get("user")
	if n != u1.Name {
		t.Error("user name should be Joe")
	}

	// Render
	tpl := &Template{
		templates: template.Must(template.New("hello").Parse("{{.}}")),
	}
	c.echo.renderer = tpl
	if he := c.Render(http.StatusOK, "hello", "Joe"); he != nil {
		t.Errorf("render %#v", he.Error)
	}
	c.echo.renderer = nil
	if he := c.Render(http.StatusOK, "hello", "Joe"); he.Error == nil {
		t.Error("render should error out")
	}

	// JSON
	r.Header.Set(Accept, ApplicationJSON)
	c.Response.committed = false
	if he := c.JSON(http.StatusOK, u1); he != nil {
		t.Errorf("json %#v", he)
	}

	// String
	r.Header.Set(Accept, TextPlain)
	c.Response.committed = false
	if he := c.String(http.StatusOK, "Hello, World!"); he != nil {
		t.Errorf("string %#v", he.Error)
	}

	// HTML
	r.Header.Set(Accept, TextHTML)
	c.Response.committed = false
	if he := c.HTML(http.StatusOK, "Hello, <strong>World!</strong>"); he != nil {
		t.Errorf("html %v", he.Error)
	}

	// Redirect
	c.Response.committed = false
	c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo")
}
