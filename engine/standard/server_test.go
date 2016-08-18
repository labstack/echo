package standard

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/stretchr/testify/assert"
)

// TODO: Fix me
func TestServer(t *testing.T) {
	s := New("")
	s.SetHandler(engine.HandlerFunc(func(req engine.Request, res engine.Response) {
	}))
	rec := httptest.NewRecorder()
	req := new(http.Request)
	s.ServeHTTP(rec, req)
}

func TestServerWrapHandler(t *testing.T) {
	e := echo.New()
	req := NewRequest(new(http.Request), nil)
	rec := httptest.NewRecorder()
	res := NewResponse(rec, nil)
	c := e.NewContext(req, res)
	h := WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test", rec.Body.String())
	}
}

func TestServerWrapMiddleware(t *testing.T) {
	e := echo.New()
	req := NewRequest(new(http.Request), nil)
	rec := httptest.NewRecorder()
	res := NewResponse(rec, nil)
	c := e.NewContext(req, res)
	buf := new(bytes.Buffer)
	mw := WrapMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf.Write([]byte("mw"))
			h.ServeHTTP(w, r)
		})
	})
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, "mw", buf.String())
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	}
}
