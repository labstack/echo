package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestStripTrailingSlash(t *testing.T) {
	req, _ := http.NewRequest(echo.GET, "/users/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec), echo.New())
	StripTrailingSlash()(c)
	assert.Equal(t, "/users", c.Request().URL.Path)
}
