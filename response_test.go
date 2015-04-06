package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponse(t *testing.T) {
	e := New()
	e.Get("/hello", func(c *Context) {
		c.String(http.StatusOK, "world")

		// Status
		if c.Response.Status() != http.StatusOK {
			t.Error("status code should be 200")
		}

		// Size
		if c.Response.Status() != http.StatusOK {
			t.Error("size should be 5")
		}

		// TODO: fix us later
		c.Response.CloseNotify()
		c.Response.Flusher()
		c.Response.Hijack()

		// Reset
		c.Response.reset(c.Response.ResponseWriter)
	})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/hello", nil)
	e.ServeHTTP(w, r)
}
