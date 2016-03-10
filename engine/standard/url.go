package standard

import "net/url"

type (
	URL struct {
		*url.URL
		query url.Values
	}
)

func (u *URL) SetPath(path string) {
	u.URL.Path = path
}

func (u *URL) Path() string {
	return u.URL.Path
}

func (u *URL) QueryValue(name string) string {
	if u.query == nil {
		u.query = u.Query()
	}
	return u.query.Get(name)
}

func (u *URL) reset(url *url.URL) {
	u.URL = url
}
