package echo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContext(t *testing.T) {
	e := New()
	e.Put("/users/:id", func(c *Context) {
		u := new(user)

		// Param
		if c.Param("id") != "1" {
			t.Error("param by name, id should be 1")
		}
		if c.P(0) != "1" {
			t.Error("param by index, id should be 1")
		}

		// Store
		c.Set("user", u.Name)
		n := c.Get("user")
		if n != u.Name {
			t.Error("user name should be Joe")
		}

		// Bind & JSON
		if err := c.Bind(u); err == nil {
			c.JSON(http.StatusCreated, u)
		}

		// TODO: fix me later
		c.Redirect(http.StatusMovedPermanently, "")
	})

	b, _ := json.Marshal(u)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(MethodPUT, "/users/1", bytes.NewReader(b))
	r.Header.Add(HeaderContentType, MIMEJSON)
	e.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Errorf("status code should be 201, found %d", w.Code)
	}
	verifyUser(w.Body, t)
}
