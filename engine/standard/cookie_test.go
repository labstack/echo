package standard

import (
	"github.com/labstack/echo/engine/test"
	"net/http"
	"testing"
	"time"
)

func TestCookie(t *testing.T) {
	cookie := &Cookie{&http.Cookie{
		Name:     "session",
		Value:    "securetoken",
		Path:     "/",
		Domain:   "github.com",
		Expires:  time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC),
		Secure:   true,
		HttpOnly: true,
	}}
	test.CookieTest(t, cookie)
}
