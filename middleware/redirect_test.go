package middleware

import (
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestHTTPSRedirect(t *testing.T) {
	e := echo.New()
	next := func(c echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := test.NewRequest(echo.GET, "http://labstack.com", nil)
	res := test.NewResponseRecorder()
	c := e.NewContext(req, res)
	HTTPSRedirect()(next)(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Status())
	assert.Equal(t, "https://labstack.com", res.Header().Get(echo.HeaderLocation))
}

func TestHTTPSWWWRedirect(t *testing.T) {
	e := echo.New()
	next := func(c echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := test.NewRequest(echo.GET, "http://labstack.com", nil)
	res := test.NewResponseRecorder()
	c := e.NewContext(req, res)
	HTTPSWWWRedirect()(next)(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Status())
	assert.Equal(t, "https://www.labstack.com", res.Header().Get(echo.HeaderLocation))
}

func TestWWWRedirect(t *testing.T) {
	e := echo.New()
	next := func(c echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := test.NewRequest(echo.GET, "http://labstack.com", nil)
	res := test.NewResponseRecorder()
	c := e.NewContext(req, res)
	WWWRedirect()(next)(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Status())
	assert.Equal(t, "http://www.labstack.com", res.Header().Get(echo.HeaderLocation))
}

func TestNonWWWRedirect(t *testing.T) {
	e := echo.New()
	next := func(c echo.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := test.NewRequest(echo.GET, "http://www.labstack.com", nil)
	res := test.NewResponseRecorder()
	c := e.NewContext(req, res)
	NonWWWRedirect()(next)(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Status())
	assert.Equal(t, "http://labstack.com", res.Header().Get(echo.HeaderLocation))
}
