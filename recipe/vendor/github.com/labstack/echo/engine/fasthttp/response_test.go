package fasthttp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/labstack/gommon/log"
)

func TestResponseWriteHeader(t *testing.T) {
	c := new(fasthttp.RequestCtx)
	res := NewResponse(c, log.New("test"))
	res.WriteHeader(http.StatusOK)
	assert.True(t, res.Committed())
	assert.Equal(t, http.StatusOK, res.Status())
}

func TestResponseWrite(t *testing.T) {
	c := new(fasthttp.RequestCtx)
	res := NewResponse(c, log.New("test"))
	res.Write([]byte("test"))
	assert.Equal(t, int64(4), res.Size())
	assert.Equal(t, "test", string(c.Response.Body()))
}

func TestResponseSetCookie(t *testing.T) {
	c := new(fasthttp.RequestCtx)
	res := NewResponse(c, log.New("test"))
	cookie := new(fasthttp.Cookie)
	cookie.SetKey("name")
	cookie.SetValue("Jon Snow")
	res.SetCookie(&Cookie{cookie})
	c.Response.Header.SetCookie(cookie)
	ck := new(fasthttp.Cookie)
	ck.SetKey("name")
	assert.True(t, c.Response.Header.Cookie(ck))
	assert.Equal(t, "Jon Snow", string(ck.Value()))
}
