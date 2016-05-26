// +build !appengine

package fasthttp

import "github.com/valyala/fasthttp"

type (
	// URL implements `engine.URL`.
	URL struct {
		*fasthttp.URI
	}
)

// Scheme implements `engine.URL#Scheme` function.
func (u *URL) Scheme() string {
	return string(u.URI.Scheme())
}

// Path implements `engine.URL#Path` function.
func (u *URL) Path() string {
	return string(u.URI.Path())
}

// SetPath implements `engine.URL#SetPath` function.
func (u *URL) SetPath(path string) {
	u.URI.SetPath(path)
}

// QueryParam implements `engine.URL#QueryParam` function.
func (u *URL) QueryParam(name string) string {
	return string(u.QueryArgs().Peek(name))
}

// QueryParams implements `engine.URL#QueryParams` function.
func (u *URL) QueryParams() (params map[string][]string) {
	params = make(map[string][]string)
	u.QueryArgs().VisitAll(func(k, v []byte) {
		// TODO: Filling with only first value
		params[string(k)] = []string{string(v)}
	})
	return
}

// QueryString implements `engine.URL#QueryString` function.
func (u *URL) QueryString() string {
	return string(u.URI.QueryString())
}

func (u *URL) reset(uri *fasthttp.URI) {
	u.URI = uri
}
