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
	rq := test.NewRequest(echo.GET, "/add-slash", nil)
	rc := test.NewResponseRecorder()
	c := echo.NewContext(rq, rc, e)
	h := AddTrailingSlash()(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, "/add-slash/", rq.URL().Path())
	assert.Equal(t, "/add-slash/", rq.URI())

	// With config
	rq = test.NewRequest(echo.GET, "/add-slash?key=value", nil)
	rc = test.NewResponseRecorder()
	c = echo.NewContext(rq, rc, e)
	h = AddTrailingSlashWithConfig(TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	})(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, http.StatusMovedPermanently, rc.Status())
	assert.Equal(t, "/add-slash/?key=value", rc.Header().Get(echo.HeaderLocation))
}

func TestRemoveTrailingSlash(t *testing.T) {
	e := echo.New()
	rq := test.NewRequest(echo.GET, "/remove-slash/", nil)
	rc := test.NewResponseRecorder()
	c := echo.NewContext(rq, rc, e)
	h := RemoveTrailingSlash()(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, "/remove-slash", rq.URL().Path())
	assert.Equal(t, "/remove-slash", rq.URI())

	// With config
	rq = test.NewRequest(echo.GET, "/remove-slash/?key=value", nil)
	rc = test.NewResponseRecorder()
	c = echo.NewContext(rq, rc, e)
	h = RemoveTrailingSlashWithConfig(TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	})(func(c echo.Context) error {
		return nil
	})
	h(c)
	assert.Equal(t, http.StatusMovedPermanently, rc.Status())
	assert.Equal(t, "/remove-slash?key=value", rc.Header().Get(echo.HeaderLocation))
}
