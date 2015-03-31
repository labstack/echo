package echo

import (
	"net/http/httptest"
	"testing"
)

func TestContextBind(t *testing.T) {
}

func TestContextJSON(t *testing.T) {
}

func TestContextHTML(t *testing.T) {
	res := httptest.NewRecorder()

	c := Context{Response: &response{}}
	c.Response.reset(res)
	c.HTML(201, "hello")

	if res.Code != 201 {
		t.Errorf("status code should be 200, found %d", res.Code)
	}

	ct := res.Header().Get(HeaderContentType)
	ex := "text/html; charset=utf-8"
	if ct != ex {
		t.Errorf("context type should be %s, found %s", ex, ct)
	}

	body := res.Body.String()
	if body != "hello" {
		t.Errorf("body should be 'hello', found %s", body)
	}
}
