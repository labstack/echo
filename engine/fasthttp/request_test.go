package fasthttp

import (
	"bufio"
	"bytes"
	"github.com/labstack/echo/engine/test"
	"github.com/labstack/gommon/log"
	fast "github.com/valyala/fasthttp"
	"net"
	"net/url"
	"testing"
)

type fakeAddr struct {
	addr string
	net.Addr
}

func (a fakeAddr) String() string {
	return a.addr
}

func TestRequest(t *testing.T) {
	var ctx fast.RequestCtx

	url, _ := url.Parse("https://github.com/labstack/echo")
	ctx.Init(&fast.Request{}, fakeAddr{addr: "127.0.0.1"}, nil)
	ctx.Request.Read(bufio.NewReader(bytes.NewBufferString(test.MultipartRequest)))
	ctx.Request.SetRequestURI(url.String())

	test.RequestTest(t, NewRequest(&ctx, log.New("echo")))
}
