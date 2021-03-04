package middleware

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestRewritePath(t *testing.T) {
	var testCases = []struct {
		whenURL       string
		expectPath    string
		expectRawPath string
	}{
		{
			whenURL:       "http://localhost:8080/old",
			expectPath:    "/new",
			expectRawPath: "",
		},
		{ // encoded `ol%64` (decoded `old`) should not be rewritten to `/new`
			whenURL:       "/ol%64", // `%64` is decoded `d`
			expectPath:    "/old",
			expectRawPath: "/ol%64",
		},
		{
			whenURL:       "http://localhost:8080/users/+_+/orders/___++++?test=1",
			expectPath:    "/user/+_+/order/___++++",
			expectRawPath: "",
		},
		{
			whenURL:       "http://localhost:8080/users/%20a/orders/%20aa",
			expectPath:    "/user/ a/order/ aa",
			expectRawPath: "",
		},
		{
			whenURL:       "http://localhost:8080/%47%6f%2f",
			expectPath:    "/Go/",
			expectRawPath: "/%47%6f%2f",
		},
		{
			whenURL:       "/users/jill/orders/T%2FcO4lW%2Ft%2FVp%2F",
			expectPath:    "/user/jill/order/T/cO4lW/t/Vp/",
			expectRawPath: "/user/jill/order/T%2FcO4lW%2Ft%2FVp%2F",
		},
		{ // do nothing, replace nothing
			whenURL:       "http://localhost:8080/user/jill/order/T%2FcO4lW%2Ft%2FVp%2F",
			expectPath:    "/user/jill/order/T/cO4lW/t/Vp/",
			expectRawPath: "/user/jill/order/T%2FcO4lW%2Ft%2FVp%2F",
		},
	}

	rules := map[*regexp.Regexp]string{
		regexp.MustCompile("^/old$"):                      "/new",
		regexp.MustCompile("^/users/(.*?)/orders/(.*?)$"): "/user/$1/order/$2",
	}

	for _, tc := range testCases {
		t.Run(tc.whenURL, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.whenURL, nil)

			rewritePath(rules, req)

			assert.Equal(t, tc.expectPath, req.URL.Path)       // Path field is stored in decoded form: /%47%6f%2f becomes /Go/.
			assert.Equal(t, tc.expectRawPath, req.URL.RawPath) // RawPath, an optional field which only gets set if the default encoding is different from Path.
		})
	}
}
