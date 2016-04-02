package middleware

import (
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
}
