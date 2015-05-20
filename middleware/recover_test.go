package middleware

import (
	"github.com/labstack/echo"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecover(t *testing.T) {
	e := echo.New()
	e.SetDebug(true)
	req, _ := http.NewRequest(echo.GET, "/", nil)
	w := httptest.NewRecorder()
	res := &echo.Response{Writer: w}
	c := echo.NewContext(req, res, e)
	h := func(c *echo.Context) *echo.HTTPError {
		panic("test")
	}

	// Status
	Recover()(h)(c)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status `500`, got %d.", w.Code)
	}

	// Body
	s := w.Body.String()
	if !strings.Contains(s, "panic recover") {
		t.Error("expected body contains `panice recover`.")
	}
}
