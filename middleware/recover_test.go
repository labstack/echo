package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	e := echo.New()
	var stack []byte
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	rc := RecoverConfig{
		RecoverHandler: func(c echo.Context, err error, s []byte) {
			stack = s
		},
	}
	h := RecoverWithConfig(rc)(echo.HandlerFunc(func(c echo.Context) error {
		panic("test")
	}))
	h(c)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, string(stack), "panic")
}
