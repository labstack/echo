package middleware

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	e := echo.New()
	e.SetDebug(true)
	buf := new(bytes.Buffer)
	e.SetLogOutput(buf)
	req := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := echo.NewContext(req, rec, e)
	h := Recover()(echo.HandlerFunc(func(c echo.Context) error {
		panic("test")
	}))
	h.Handle(c)
	assert.Equal(t, http.StatusInternalServerError, rec.Status())
	assert.Contains(t, buf.String(), "PANIC RECOVER")
}
