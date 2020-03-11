package fasthttp

import (
	"github.com/labstack/echo/engine/test"
	fast "github.com/valyala/fasthttp"
	"testing"
	"time"
)

func TestCookie(t *testing.T) {
	fCookie := &fast.Cookie{}
	fCookie.SetKey("session")
	fCookie.SetValue("securetoken")
	fCookie.SetPath("/")
	fCookie.SetDomain("github.com")
	fCookie.SetExpire(time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC))
	fCookie.SetSecure(true)
	fCookie.SetHTTPOnly(true)

	cookie := &Cookie{
		fCookie,
	}
	test.CookieTest(t, cookie)
}
