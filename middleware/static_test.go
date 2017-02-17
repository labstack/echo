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
	// TODO: Once go1.6 is dropped, use `httptest.Request()`.
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("*")
	config := StaticConfig{
		Root: "../_fixture",
	}

	// Directory
	h := StaticWithConfig(config)(echo.NotFoundHandler)
	if assert.NoError(t, h(c)) {
		assert.Contains(t, rec.Body.String(), "Echo")
	}

	// File found
	h = StaticWithConfig(config)(echo.NotFoundHandler)
	c.SetParamValues("/images/walle.png")
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, rec.Header().Get(echo.HeaderContentLength), "219885")
	}

	// File not found
	c.SetParamValues("/none")
	rec.Body.Reset()
	he := h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusNotFound, he.Code)

	// HTML5
	c.SetParamValues("/random")
	rec.Body.Reset()
	config.HTML5 = true
	static := StaticWithConfig(config)
	h = static(echo.NotFoundHandler)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Echo")
	}

	// Browse
	c.SetParamValues("/")
	rec.Body.Reset()
	config.Browse = true
	static = StaticWithConfig(config)
	h = static(echo.NotFoundHandler)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "images")
	}
}
