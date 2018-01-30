package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

type middlewareGenerator func() echo.MiddlewareFunc

func TestRedirectHTTPSRedirect(t *testing.T) {
	res := redirectTest(HTTPSRedirect, "labstack.com")

	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://labstack.com/", res.Header().Get(echo.HeaderLocation))
}

func TestRedirectHTTPSWWWRedirect(t *testing.T) {
	res := redirectTest(HTTPSWWWRedirect, "labstack.com")

	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://www.labstack.com/", res.Header().Get(echo.HeaderLocation))
}

func TestRedirectHTTPSNonWWWRedirect(t *testing.T) {
	res := redirectTest(HTTPSNonWWWRedirect, "www.labstack.com")

	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://labstack.com/", res.Header().Get(echo.HeaderLocation))
}

func TestRedirectWWWRedirect(t *testing.T) {
	res := redirectTest(WWWRedirect, "labstack.com")

	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "http://www.labstack.com/", res.Header().Get(echo.HeaderLocation))
}

func TestRedirectNonWWWRedirect(t *testing.T) {
	res := redirectTest(NonWWWRedirect, "www.labstack.com")

	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "http://labstack.com/", res.Header().Get(echo.HeaderLocation))
}

func redirectTest(fn middlewareGenerator, host string) *httptest.ResponseRecorder {
	e := echo.New()
	next := func(c echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(echo.GET, "/", nil)
	req.Host = host
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)

	fn()(next)(c)

	return res
}
