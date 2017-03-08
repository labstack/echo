package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestStatic(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	config := StaticConfig{
		Root: "../_fixture",
	}

	// Directory
	h := StaticWithConfig(config)(echo.NotFoundHandler)
	if assert.NoError(t, h(c)) {
		assert.Contains(t, rec.Body.String(), "Echo")
	}

	// File found
	req = httptest.NewRequest(echo.GET, "/images/walle.png", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, rec.Header().Get(echo.HeaderContentLength), "219885")
	}

	// File not found
	req = httptest.NewRequest(echo.GET, "/none", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	he := h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusNotFound, he.Code)

	// HTML5
	req = httptest.NewRequest(echo.GET, "/random", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	config.HTML5 = true
	static := StaticWithConfig(config)
	h = static(echo.NotFoundHandler)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Echo")
	}

	// Browse
	req = httptest.NewRequest(echo.GET, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	config.Root = "../_fixture/certs"
	config.Browse = true
	static = StaticWithConfig(config)
	h = static(echo.NotFoundHandler)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "cert.pem")
	}
}
