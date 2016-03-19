// +build !appengine

package fasthttp

import "github.com/valyala/fasthttp"

type (
	// URL implements `engine.URL`.
	URL struct {
		*fasthttp.URI
	}
)

// Path implements `engine.URL#Path` function.
func (u *URL) Path() string {
	return string(u.URI.Path())
}

// SetPath implements `engine.URL#SetPath` function.
func (u *URL) SetPath(path string) {
	u.URI.SetPath(path)
}

// QueryValue implements `engine.URL#QueryValue` function.
func (u *URL) QueryValue(name string) string {
	return string(u.QueryArgs().Peek(name))
}

// QueryString implements `engine.URL#QueryString` function.
func (u *URL) QueryString() string {
	return string(u.URI.QueryString())
}

func (u *URL) reset(uri *fasthttp.URI) {
	u.URI = uri
}
