// +build !appengine

package fasthttp

import "github.com/valyala/fasthttp"

type (
	// URL implements `engine.URL`.
	URL struct {
		*fasthttp.URI
	}
)

// Path implements `URL#Path` method.
func (u *URL) Path() string {
	return string(u.URI.Path())
}

// SetPath implements `URL#SetPath` method.
func (u *URL) SetPath(path string) {
	// return string(u.URI.Path())
}

// QueryValue implements `URL#QueryValue` method.
func (u *URL) QueryValue(name string) string {
	return ""
}

func (u *URL) reset(uri *fasthttp.URI) {
	u.URI = uri
}
