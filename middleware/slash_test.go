package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestStripTrailingSlash(t *testing.T) {
	req, _ := http.NewRequest(echo.GET, "/users/", nil)
	res := &echo.Response{Writer: httptest.NewRecorder()}
	c := echo.NewContext(req, res, echo.New())
	StripTrailingSlash()(c)
	if c.Request.URL.Path != "/users" {
		t.Error("it should strip the trailing slash")
	}
}

func TestRedirectToSlash(t *testing.T) {
	req, _ := http.NewRequest(echo.GET, "/users", nil)
	res := &echo.Response{Writer: httptest.NewRecorder()}
	c := echo.NewContext(req, res, echo.New())
	RedirectToSlash(RedirectToSlashOptions{Code: http.StatusTemporaryRedirect})(c)
	if res.Status() != http.StatusTemporaryRedirect {
		t.Errorf("status code should be 307, found %d", res.Status())
	}
	if c.Response.Header().Get("Location") != "/users/" {
		t.Error("Location header should be /users/")
	}
}
