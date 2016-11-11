package middleware

import (
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestStatic(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	h := Static("../_fixture")(func(c echo.Context) error {
		return echo.ErrNotFound
	})

	// Directory
	if assert.NoError(t, h(c)) {
		assert.Contains(t, rec.Body.String(), "Echo")
	}

	// HTML5 mode
	req = test.NewRequest(echo.GET, "/client", nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	static := StaticWithConfig(StaticConfig{
		Root:  "../_fixture",
		HTML5: true,
	})
	h = static(func(c echo.Context) error {
		return echo.ErrNotFound
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Status())
	}

	// Browse
	req = test.NewRequest(echo.GET, "/", nil)
	rec = test.NewResponseRecorder()
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
	req = test.NewRequest(echo.GET, "/not-found", nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	static = StaticWithConfig(StaticConfig{
		Root: "../_fixture/images",
	})
	h = static(func(c echo.Context) error {
		return echo.ErrNotFound
	})
	assert.Error(t, h(c))
}
