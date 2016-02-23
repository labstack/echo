// +build !appengine

package fasthttp

import "github.com/valyala/fasthttp"

type (
	URL struct {
		url *fasthttp.URI
	}
)

func (u *URL) SetPath(path string) {
	// return string(u.URI.Path())
}

func (u *URL) Path() string {
	return string(u.url.Path())
}

func (u *URL) QueryValue(name string) string {
	return ""
}

func (u *URL) Object() interface{} {
	return u.url
}

func (u *URL) reset(url *fasthttp.URI) {
	u.url = url
}
