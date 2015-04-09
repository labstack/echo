package echo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContext(t *testing.T) {
	b, _ := json.Marshal(u1)
	r, _ := http.NewRequest(POST, "/users/1", bytes.NewReader(b))
	c := &Context{
		Response: &response{writer: httptest.NewRecorder()},
		Request:  r,
		params:   make(Params, 5),
		store:    make(store),
	}

	//**********//
	//   Bind   //
	//**********//
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

	//***********//
	//   Param   //
	//***********//
	// By id
	c.params = Params{{"id", "1"}}
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

	// JSON
	r.Header.Set(HeaderAccept, MIMEJSON)
	if err := c.JSON(http.StatusOK, u1); err != nil {
		t.Errorf("render json %v", err)
	}

	// String
	r.Header.Set(HeaderAccept, MIMEText)
	c.Response.committed = false
	if err := c.String(http.StatusOK, "Hello, World!"); err != nil {
		t.Errorf("render string %v", err)
	}

	// HTML
	r.Header.Set(HeaderAccept, MIMEHTML)
	c.Response.committed = false
	if err := c.HTML(http.StatusOK, "Hello, <strong>World!</strong>"); err != nil {
		t.Errorf("render html %v", err)
	}

	// Redirect
	c.Response.committed = false
	c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo")
}
