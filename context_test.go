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
	c := &Context{
		Response: &response{Writer: httptest.NewRecorder()},
		Request:  r,
		pvalues:  make([]string, 5),
		store:    make(store),
		echo:     New(),
	}

	//------
	// Bind
	//------

	// JSON
	r.Header.Set(HeaderContentType, MIMEJSON)
	u2 := new(user)
	if err := c.Bind(u2); err != nil {
		t.Error(err)
	}
	verifyUser(u2, t)

	// FORM
	r.Header.Set(HeaderContentType, MIMEForm)
	u2 = new(user)
	if err := c.Bind(u2); err != nil {
		t.Error(err)
	}
	// TODO: add verification

	// Unsupported
	r.Header.Set(HeaderContentType, "")
	u2 = new(user)
	if err := c.Bind(u2); err == nil {
		t.Error(err)
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
	if err := c.Render(http.StatusOK, "hello", "Joe"); err != nil {
		t.Errorf("render %v", err)
	}
	c.echo.renderer = nil
	if err := c.Render(http.StatusOK, "hello", "Joe"); err == nil {
		t.Error("render should error out")
	}

	// JSON
	r.Header.Set(HeaderAccept, MIMEJSON)
	c.Response.committed = false
	if err := c.JSON(http.StatusOK, u1); err != nil {
		t.Errorf("json %v", err)
	}

	// String
	r.Header.Set(HeaderAccept, MIMEText)
	c.Response.committed = false
	if err := c.String(http.StatusOK, "Hello, World!"); err != nil {
		t.Errorf("string %v", err)
	}

	// HTML
	r.Header.Set(HeaderAccept, MIMEHTML)
	c.Response.committed = false
	if err := c.HTML(http.StatusOK, "Hello, <strong>World!</strong>"); err != nil {
		t.Errorf("html %v", err)
	}

	// Redirect
	c.Response.committed = false
	c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo")
}
