package fasthttp

import "github.com/valyala/fasthttp"

type (
	URL struct {
		url *fasthttp.URI
	}
)

func (u *URL) Scheme() string {
	return string(u.url.Scheme())
}

func (u *URL) Host() string {
	return string(u.url.Host())
}

func (u *URL) SetPath(path string) {
	// return string(u.URI.Path())
}

func (u *URL) Path() string {
	return string(u.url.Path())
}

func (u *URL) QueryValue(name string) string {
	return ""
}
