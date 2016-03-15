package standard

import "net/url"

type (
	// URL implements `engine.URL`.
	URL struct {
		*url.URL
		query url.Values
	}
)

// Path implements `engine.URL#Path` method.
func (u *URL) Path() string {
	return u.URL.Path
}

// SetPath implements `engine.URL#SetPath` method.
func (u *URL) SetPath(path string) {
	u.URL.Path = path
}

// QueryValue implements `engine.URL#QueryValue` method.
func (u *URL) QueryValue(name string) string {
	if u.query == nil {
		u.query = u.Query()
	}
	return u.query.Get(name)
}

// QueryString implements `engine.URL#QueryString` method.
func (u *URL) QueryString() string {
	return u.URL.RawQuery
}

func (u *URL) reset(url *url.URL) {
	u.URL = url
}
