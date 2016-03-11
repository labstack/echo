package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	echo "gopkg.in/labstack/echo.v1"
)

func TestRecover(t *testing.T) {
	e := echo.New()
	e.SetDebug(true)
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, echo.NewResponse(rec, e), e)
	h := func(c *echo.Context) error {
		panic("test")
	}
	Recover()(h)(c)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "panic recover")
}
