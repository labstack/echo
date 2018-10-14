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
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	config := StaticConfig{
		Root: "../_fixture",
	}

	// Directory
	h := StaticWithConfig(config)(echo.NotFoundHandler)

	assert := assert.New(t)

	if assert.NoError(h(c)) {
		assert.Contains(rec.Body.String(), "Echo")
	}

	// File found
	req = httptest.NewRequest(http.MethodGet, "/images/walle.png", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(h(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(rec.Header().Get(echo.HeaderContentLength), "219885")
	}

	// File not found
	req = httptest.NewRequest(http.MethodGet, "/none", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	he := h(c).(*echo.HTTPError)
	assert.Equal(http.StatusNotFound, he.Code)

	// HTML5
	req = httptest.NewRequest(http.MethodGet, "/random", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	config.HTML5 = true
	static := StaticWithConfig(config)
	h = static(echo.NotFoundHandler)
	if assert.NoError(h(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Contains(rec.Body.String(), "Echo")
	}

	// Browse
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	config.Root = "../_fixture/certs"
	config.Browse = true
	static = StaticWithConfig(config)
	h = static(echo.NotFoundHandler)
	if assert.NoError(h(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Contains(rec.Body.String(), "cert.pem")
	}
}
