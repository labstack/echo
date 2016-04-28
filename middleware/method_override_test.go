package middleware

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestOverrideMtrhod(t *testing.T) {
	e := echo.New()
	methodOverride := OverrideMethod()
	h := methodOverride(func(c echo.Context) error {
		return c.String(http.StatusOK, c.Request().Method())
	})

	// Override with http header
	rq := test.NewRequest(echo.POST, "/", nil)
	rq.Header().Set(HttpMethodOverrideHeader, "DELETE")
	rc := test.NewResponseRecorder()
	c := e.NewContext(rq, rc)
	h(c)
	assert.Equal(t, "DELETE", rc.Body.String())

	// Override with body parameter
	rq = test.NewRequest(echo.POST, "/", bytes.NewReader([]byte("_method=DELETE")))
	rq.Header().Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc)
	h(c)
	assert.Equal(t, "DELETE", rc.Body.String())

	// Ignore GET
	rq = test.NewRequest(echo.GET, "/", nil)
	rq.Header().Set(HttpMethodOverrideHeader, "DELETE")
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc)
	h(c)
	assert.Equal(t, "GET", rc.Body.String())
}
