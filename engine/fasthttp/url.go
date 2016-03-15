// +build !appengine

package fasthttp

import "github.com/valyala/fasthttp"

type (
	// URL implements `engine.URL`.
	URL struct {
		*fasthttp.URI
	}
)

// Path implements `engine.URL#Path` method.
func (u *URL) Path() string {
	return string(u.URI.Path())
}

// SetPath implements `engine.URL#SetPath` method.
func (u *URL) SetPath(path string) {
	u.URI.SetPath(path)
}

// QueryValue implements `engine.URL#QueryValue` method.
func (u *URL) QueryValue(name string) string {
	return string(u.QueryArgs().Peek(name))
}

func (u *URL) reset(uri *fasthttp.URI) {
	u.URI = uri
}
