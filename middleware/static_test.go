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
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Static("../_fixture")(func(c echo.Context) error {
		return echo.ErrNotFound
	})

	// Directory
	if assert.NoError(t, h(c)) {
		assert.Contains(t, rec.Body.String(), "Echo")
	}

	// HTML5 mode
	req, _ = http.NewRequest(echo.GET, "/client", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	static := StaticWithConfig(StaticConfig{
		Root:  "../_fixture",
		HTML5: true,
	})
	h = static(func(c echo.Context) error {
		return echo.ErrNotFound
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Browse
	req, _ = http.NewRequest(echo.GET, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	static = StaticWithConfig(StaticConfig{
		Root:   "../_fixture/images",
		Browse: true,
	})
	h = static(func(c echo.Context) error {
		return echo.ErrNotFound
	})
	if assert.NoError(t, h(c)) {
		assert.Contains(t, rec.Body.String(), "walle")
	}

	// Not found
	req, _ = http.NewRequest(echo.GET, "/not-found", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	static = StaticWithConfig(StaticConfig{
		Root: "../_fixture/images",
	})
	h = static(func(c echo.Context) error {
		return echo.ErrNotFound
	})
	assert.Error(t, h(c))
}
