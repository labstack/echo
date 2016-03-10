// +build !appengine

package fasthttp

import "github.com/valyala/fasthttp"

type (
	URL struct {
		*fasthttp.URI
	}
)

func (u *URL) SetPath(path string) {
	// return string(u.URI.Path())
}

func (u *URL) Path() string {
	return string(u.URI.Path())
}

func (u *URL) QueryValue(name string) string {
	return ""
}

func (u *URL) reset(uri *fasthttp.URI) {
	u.URI = uri
}
