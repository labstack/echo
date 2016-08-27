package middleware

import (
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestAddTrailingSlash(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/add-slash", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	h := AddTrailingSlash()(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, "/add-slash/", req.URL().Path())
	assert.Equal(t, "/add-slash/", req.URI())

	// With config
	req = test.NewRequest(echo.GET, "/add-slash?key=value", nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	h = AddTrailingSlashWithConfig(TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	})(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, http.StatusMovedPermanently, rec.Status())
	assert.Equal(t, "/add-slash/?key=value", rec.Header().Get(echo.HeaderLocation))
}

func TestRemoveTrailingSlash(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/remove-slash/", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	h := RemoveTrailingSlash()(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, "/remove-slash", req.URL().Path())
	assert.Equal(t, "/remove-slash", req.URI())

	// With config
	req = test.NewRequest(echo.GET, "/remove-slash/?key=value", nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	h = RemoveTrailingSlashWithConfig(TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	})(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, http.StatusMovedPermanently, rec.Status())
	assert.Equal(t, "/remove-slash?key=value", rec.Header().Get(echo.HeaderLocation))

	// With bare URL
	req = test.NewRequest(echo.GET, "http://localhost", nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	h = RemoveTrailingSlash()(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, "", req.URL().Path())
}
