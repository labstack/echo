// +build go1.7, !go1.8

package echo

import (
	"net/url"
)

// PathUnescape is wraps `url.QueryUnescape`
func PathUnescape(s string) (string, error) {
	return url.QueryUnescape(s)
}
