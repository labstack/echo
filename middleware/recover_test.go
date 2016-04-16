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
	buf := new(bytes.Buffer)
	e.SetLogOutput(buf)
	rq := test.NewRequest(echo.GET, "/", nil)
	rc := test.NewResponseRecorder()
	c := e.NewContext(rq, rc)
	h := Recover()(echo.HandlerFunc(func(c echo.Context) error {
		panic("test")
	}))
	h(c)
	assert.Equal(t, http.StatusInternalServerError, rc.Status())
	assert.Contains(t, buf.String(), "PANIC RECOVER")
}
