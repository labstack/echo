package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	res := &Response{echo: e, Writer: rec}

	// Before
	res.Before(func() {
		c.Response().Header().Set(HeaderServer, "echo")
	})
	res.Write([]byte("test"))
	assert.Equal(t, "echo", rec.Header().Get(HeaderServer))
}
