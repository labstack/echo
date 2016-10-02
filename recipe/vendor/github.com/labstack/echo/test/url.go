package test

import "net/url"

type (
	URL struct {
		url   *url.URL
		query url.Values
	}
)

func (u *URL) URL() *url.URL {
	return u.url
}

func (u *URL) SetPath(path string) {
	u.url.Path = path
}

func (u *URL) Path() string {
	return u.url.Path
}

func (u *URL) QueryParam(name string) string {
	if u.query == nil {
		u.query = u.url.Query()
	}
	return u.query.Get(name)
}

func (u *URL) QueryParams() map[string][]string {
	if u.query == nil {
		u.query = u.url.Query()
	}
	return map[string][]string(u.query)
}

func (u *URL) QueryString() string {
	return u.url.RawQuery
}

func (u *URL) reset(url *url.URL) {
	u.url = url
}
