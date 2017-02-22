package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestRedirectHTTPSRedirect(t *testing.T) {
	e := echo.New()
	next := func(c echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(echo.GET, "/", nil)
	req.Host = "labstack.com"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	HTTPSRedirect()(next)(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://labstack.com/", res.Header().Get(echo.HeaderLocation))
}

func TestRedirectHTTPSWWWRedirect(t *testing.T) {
	e := echo.New()
	next := func(c echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(echo.GET, "/", nil)
	req.Host = "labstack.com"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	HTTPSWWWRedirect()(next)(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://www.labstack.com/", res.Header().Get(echo.HeaderLocation))
}

func TestRedirectHTTPSNonWWWRedirect(t *testing.T) {
	e := echo.New()
	next := func(c echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(echo.GET, "/", nil)
	req.Host = "www.labstack.com"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	HTTPSNonWWWRedirect()(next)(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://labstack.com/", res.Header().Get(echo.HeaderLocation))
}

func TestRedirectWWWRedirect(t *testing.T) {
	e := echo.New()
	next := func(c echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(echo.GET, "/", nil)
	req.Host = "labstack.com"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	WWWRedirect()(next)(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "http://www.labstack.com/", res.Header().Get(echo.HeaderLocation))
}

func TestRedirectNonWWWRedirect(t *testing.T) {
	e := echo.New()
	next := func(c echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(echo.GET, "/", nil)
	req.Host = "www.labstack.com"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	NonWWWRedirect()(next)(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "http://labstack.com/", res.Header().Get(echo.HeaderLocation))
}
