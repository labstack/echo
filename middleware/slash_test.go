package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestAddTrailingSlash(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/add-slash", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := AddTrailingSlash()(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, "/add-slash/", req.URL.Path)
	assert.Equal(t, "/add-slash/", req.RequestURI)

	// With config
	req, _ = http.NewRequest(echo.GET, "/add-slash?key=value", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = AddTrailingSlashWithConfig(TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	})(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "/add-slash/?key=value", rec.Header().Get(echo.HeaderLocation))
}

func TestRemoveTrailingSlash(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/remove-slash/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := RemoveTrailingSlash()(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, "/remove-slash", req.URL.Path)
	assert.Equal(t, "/remove-slash", req.RequestURI)

	// With config
	req, _ = http.NewRequest(echo.GET, "/remove-slash/?key=value", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = RemoveTrailingSlashWithConfig(TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	})(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "/remove-slash?key=value", rec.Header().Get(echo.HeaderLocation))

	// With bare URL
	req, _ = http.NewRequest(echo.GET, "http://localhost", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = RemoveTrailingSlash()(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, "", req.URL.Path)
}
