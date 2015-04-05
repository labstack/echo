package echo

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContext(t *testing.T) {
	b, _ := json.Marshal(u1)
	r, _ := http.NewRequest(MethodPOST, "/users/1", bytes.NewReader(b))
	c := &Context{
		Response: &response{ResponseWriter: httptest.NewRecorder()},
		Request:  r,
		params:   make(Params, 5),
		store:    make(store),
	}

	//**********//
	//   Bind   //
	//**********//
	r.Header.Add(HeaderContentType, MIMEJSON)
	u2 := new(user)
	if err := c.Bind(u2); err != nil {
		t.Error(err)
	}
	verifyUser(u2, t)

	//***********//
	//   Param   //
	//***********//
	// By id
	if c.P(0) != "" {
		t.Error("param id should be nil")
	}

	// By name
	if c.Param("id") != "" {
		t.Error("param id should be nil")
	}

	// Store
	c.Set("user", u1.Name)
	n := c.Get("user")
	if n != u1.Name {
		t.Error("user name should be Joe")
	}

	//************//
	//   Render   //
	//************//
	// JSON
	r.Header.Set(HeaderAccept, MIMEJSON)
	if err := c.Render(http.StatusOK, u1); err != nil {
		t.Errorf("render json %v", err)
	}

	// String
	r.Header.Set(HeaderAccept, MIMEText)
	c.Response.committed = false
	if err := c.Render(http.StatusOK, "Hello, World!"); err != nil {
		t.Errorf("render string %v", err)
	}

	// HTML
	r.Header.Set(HeaderAccept, MIMEHTML)
	c.Response.committed = false
	if err := c.Render(http.StatusOK, "Hello, <strong>World!</strong>"); err != nil {
		t.Errorf("render html %v", err)
	}

	// HTML template
	c.Response.committed = false
	tmpl, _ := template.New("foo").Parse(`{{define "T"}}Hello, {{.}}!{{end}}`)
	if err := c.HTMLTemplate(http.StatusOK, tmpl, "T", "Joe"); err != nil {
		t.Errorf("render html template %v", err)
	}

	// Redirect
	c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo")
}
