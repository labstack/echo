package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestStripTrailingSlash(t *testing.T) {
	req, _ := http.NewRequest(echo.GET, "/users/", nil)
	res := echo.NewResponse(httptest.NewRecorder())
	c := echo.NewContext(req, res, echo.New())
	StripTrailingSlash()(c)
	p := c.Request.URL.Path
	if p != "/users" {
		t.Errorf("expected path `/users` got, %s.", p)
	}
}

func TestRedirectToSlash(t *testing.T) {
	req, _ := http.NewRequest(echo.GET, "/users", nil)
	res := echo.NewResponse(httptest.NewRecorder())
	c := echo.NewContext(req, res, echo.New())
	RedirectToSlash(RedirectToSlashOptions{Code: http.StatusTemporaryRedirect})(c)

	// Status code
	if res.Status() != http.StatusTemporaryRedirect {
		t.Errorf("expected status `307`, got %d.", res.Status())
	}

	// Location header
	l := c.Response.Header().Get("Location")
	if l != "/users/" {
		t.Errorf("expected Location header `/users/`, got %s.", l)
	}
}
