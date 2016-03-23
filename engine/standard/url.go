package standard

import "net/url"

type (
	// URL implements `engine.URL`.
	URL struct {
		*url.URL
		query url.Values
	}
)

// Path implements `engine.URL#Path` function.
func (u *URL) Path() string {
	return u.URL.Path
}

// SetPath implements `engine.URL#SetPath` function.
func (u *URL) SetPath(path string) {
	u.URL.Path = path
}

// QueryParam implements `engine.URL#QueryParam` function.
func (u *URL) QueryParam(name string) string {
	if u.query == nil {
		u.query = u.Query()
	}
	return u.query.Get(name)
}

// QueryParams implements `engine.URL#QueryParams` function.
func (u *URL) QueryParams() map[string][]string {
	if u.query == nil {
		u.query = u.Query()
	}
	return map[string][]string(u.query)
}

// QueryString implements `engine.URL#QueryString` function.
func (u *URL) QueryString() string {
	return u.URL.RawQuery
}

func (u *URL) reset(url *url.URL) {
	u.URL = url
	u.query = nil
}
