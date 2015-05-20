package middleware

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestGzip(t *testing.T) {
	// Empty Accept-Encoding header
	req, _ := http.NewRequest(echo.GET, "/", nil)
	w := httptest.NewRecorder()
	res := echo.NewResponse(w)
	c := echo.NewContext(req, res, echo.New())
	h := func(c *echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	Gzip()(h)(c)
	s := w.Body.String()
	if s != "test" {
		t.Errorf("expected `test`, with empty Accept-Encoding header, got %s.", s)
	}

	// Content-Encoding header
	req.Header.Set(echo.AcceptEncoding, "gzip")
	w = httptest.NewRecorder()
	c.Response = echo.NewResponse(w)
	Gzip()(h)(c)
	ce := w.Header().Get(echo.ContentEncoding)
	if ce != "gzip" {
		t.Errorf("expected Content-Encoding header `gzip`, got %d.", ce)
	}

	// Body
	r, err := gzip.NewReader(w.Body)
	defer r.Close()
	if err != nil {
		t.Error(err)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	s = string(b)
	if s != "test" {
		t.Errorf("expected body `test`, got %s.", s)
	}
}
