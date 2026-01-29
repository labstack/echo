package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func Test_getScheme(t *testing.T) {
	tests := []struct {
		name       string
		r          *http.Request
		headerName string
		whenHeader string
		want       string
	}{
		{
			name:       "test only X-Forwarded-Proto: https",
			headerName: "X-Forwarded-Proto",
			whenHeader: "https",
			want:       "https",
		},
		{
			name:       "test only X-Forwarded-Proto: http",
			headerName: "X-Forwarded-Proto",
			whenHeader: "http",
			want:       "http",
		},
		{
			name:       "test only X-Forwarded-Proto: HTTP",
			headerName: "X-Forwarded-Proto",
			whenHeader: "HTTP",
			want:       "http",
		},
		{
			name:       "test only X-Forwarded-Protocol: https",
			headerName: "X-Forwarded-Protocol",
			whenHeader: "https",
			want:       "https",
		},
		{
			name:       "test only X-Forwarded-Protocol: http",
			headerName: "X-Forwarded-Protocol",
			whenHeader: "http",
			want:       "http",
		},
		{
			name:       "test only X-Forwarded-Protocol: HTTP",
			headerName: "X-Forwarded-Protocol",
			whenHeader: "HTTP",
			want:       "http",
		},
		{
			name:       "test only Forwarded https",
			headerName: "Forwarded",
			whenHeader: "proto=https",
			want:       "https",
		},
		{
			name:       "test only Forwarded: http",
			headerName: "Forwarded",
			whenHeader: "proto=http",
			want:       "http",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Header: http.Header{
					tt.headerName: []string{tt.whenHeader},
				},
			}

			if got := getScheme(req); got != tt.want {
				t.Errorf("getScheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxyHeaders(t *testing.T) {
	tests := []struct {
		name        string
		givenMW     echo.MiddlewareFunc
		whenMethod  string
		whenHeaders map[string]string
		expectURL   string
	}{
		{
			name:        "Test X-Forwarded-Proto_HTTPS",
			whenMethod:  "GET",
			whenHeaders: map[string]string{echo.HeaderXForwardedProto: "HTTPS"},
			expectURL:   "https://srv.lan/tst/",
		},
		{
			name:        "Test X-Forwarded-Prefix_TEST",
			whenMethod:  "GET",
			whenHeaders: map[string]string{echo.HeaderXForwardedPrefix: "/test/"},
			expectURL:   "http://srv.lan/test/tst/",
		},
		{
			name:        "Test X-Forwarded-Prefix_TEST2",
			whenMethod:  "GET",
			whenHeaders: map[string]string{echo.HeaderXForwardedPrefix: "/test2/"},
			expectURL:   "http://srv.lan/test2/tst/",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			mw := ProxyHeaders()
			if tc.givenMW != nil {
				mw = tc.givenMW
			}

			e.Use(mw)

			h := mw(func(c echo.Context) error {
				return nil
			})

			method := http.MethodGet
			if tc.whenMethod != "" {
				method = tc.whenMethod
			}
			req := httptest.NewRequest(method, "http://srv.lan/tst/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			for k, v := range tc.whenHeaders {
				req.Header.Set(k, v)
			}

			err := h(c)

			assert.NoError(t, err)
			url := c.Request().URL.String()
			assert.Equal(t, tc.expectURL, url, "url: `%v` should be `%v`", tc.expectURL, url)
		})
	}
}
