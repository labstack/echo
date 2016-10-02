package fasthttp

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

// TODO: Fix me
func TestServer(t *testing.T) {
	s := New("")
	s.SetHandler(engine.HandlerFunc(func(req engine.Request, res engine.Response) {
	}))
	ctx := new(fasthttp.RequestCtx)
	s.ServeHTTP(ctx)
}

func TestServerWrapHandler(t *testing.T) {
	e := echo.New()
	ctx := new(fasthttp.RequestCtx)
	req := NewRequest(ctx, nil)
	res := NewResponse(ctx, nil)
	c := e.NewContext(req, res)
	h := WrapHandler(func(ctx *fasthttp.RequestCtx) {
		ctx.Write([]byte("test"))
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, ctx.Response.StatusCode())
		assert.Equal(t, "test", string(ctx.Response.Body()))
	}
}

func TestServerWrapMiddleware(t *testing.T) {
	e := echo.New()
	ctx := new(fasthttp.RequestCtx)
	req := NewRequest(ctx, nil)
	res := NewResponse(ctx, nil)
	c := e.NewContext(req, res)
	buf := new(bytes.Buffer)
	mw := WrapMiddleware(func(h fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			buf.Write([]byte("mw"))
			h(ctx)
		}
	})
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, "mw", buf.String())
		assert.Equal(t, http.StatusOK, ctx.Response.StatusCode())
		assert.Equal(t, "OK", string(ctx.Response.Body()))
	}
}
