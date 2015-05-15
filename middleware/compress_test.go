package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"compress/gzip"
	"github.com/labstack/echo"
	"io/ioutil"
)

func TestGzip(t *testing.T) {
	req, _ := http.NewRequest(echo.GET, "/", nil)
	req.Header.Set(echo.AcceptEncoding, "gzip")
	w := httptest.NewRecorder()
	res := &echo.Response{Writer: w}
	c := echo.NewContext(req, res, echo.New())
	Gzip()(func(c *echo.Context) *echo.HTTPError {
		return c.String(http.StatusOK, "test")
	})(c)

	if w.Header().Get(echo.ContentEncoding) != "gzip" {
		t.Errorf("expected Content-Encoding: gzip, got %d", w.Header().Get(echo.ContentEncoding))
	}

	r, err := gzip.NewReader(w.Body)
	defer r.Close()
	if err != nil {
		t.Error(err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	s := string(b)

	if s != "test" {
		t.Errorf(`expected "test", got "%s"`, s)
	}
}
