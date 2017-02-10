package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	e := echo.New()
	req, _ := http.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	rid := RequestIDWithConfig(RequestIDConfig{})
	h := rid(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	h(c)
	if assert.NotEmpty(t, rec.Header().Get("X-Request-ID")) {
		u, err := uuid.FromString(rec.Header().Get("X-Request-ID"))
		spew.Dump(u)
		assert.NoError(t, err)
		assert.Equal(t, uint(4), u.Version())
		assert.Equal(t, uint(uuid.VariantRFC4122), u.Variant())
	}
}
