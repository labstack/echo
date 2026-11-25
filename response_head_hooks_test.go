// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouterAutoHandleHEAD_BeforeSniffContentType(t *testing.T) {
	for _, tc := range []struct {
		name         string
		hook         func(*Response)
		wantType     string
		wantEncoding string
		wantStatus   int
	}{
		{name: "conditional default", hook: func(r *Response) {
			if r.Header().Get(HeaderContentType) == "" {
				r.Header().Set(HeaderContentType, "application/custom")
			}
		}, wantType: "application/custom"},
		{name: "unconditional override", hook: func(r *Response) { r.Header().Set(HeaderContentType, "application/custom") }, wantType: "application/custom"},
		{name: "suppression", hook: func(r *Response) { r.Header()[HeaderContentType] = nil }},
		{name: "encoding", hook: func(r *Response) { r.Header().Set(HeaderContentEncoding, "gzip") }, wantEncoding: "gzip"},
		{name: "no type change", hook: func(r *Response) { r.Header().Set("X-Hook", "ran") }, wantType: "text/html; charset=utf-8"},
		{name: "bodyless status", hook: func(r *Response) { r.Status = http.StatusNoContent }, wantStatus: http.StatusNoContent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewWithConfig(Config{Router: NewRouter(RouterConfig{AutoHandleHEAD: true})})
			calls := 0
			e.GET("/", func(c *Context) error {
				res, err := UnwrapResponse(c.Response())
				if err != nil {
					return err
				}
				res.Before(func() { calls++; tc.hook(res) })
				_, err = io.WriteString(c.Response(), "<html>hello</html>")
				return err
			})
			server := httptest.NewServer(e)
			defer server.Close()
			client := server.Client()
			client.Transport.(*http.Transport).DisableCompression = true
			for i, method := range []string{http.MethodGet, http.MethodHead} {
				req, err := http.NewRequest(method, server.URL, nil)
				require.NoError(t, err)
				res, err := client.Do(req)
				require.NoError(t, err)
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.NoError(t, res.Body.Close())
				require.Equal(t, tc.wantType, res.Header.Get(HeaderContentType), method)
				require.Equal(t, tc.wantEncoding, res.Header.Get(HeaderContentEncoding), method)
				require.Equal(t, max(http.StatusOK, tc.wantStatus), res.StatusCode, method)
				require.Equal(t, i+1, calls, method)
				if method == http.MethodHead {
					require.Empty(t, body)
				}
			}
		})
	}
}

func TestRouterAutoHandleHEAD_BeforeHookOrder(t *testing.T) {
	e := NewWithConfig(Config{Router: NewRouter(RouterConfig{AutoHandleHEAD: true})})
	e.GET("/", func(c *Context) error {
		res, err := UnwrapResponse(c.Response())
		if err != nil {
			return err
		}
		res.Before(func() { res.Header().Set(HeaderContentType, "application/first") })
		res.Before(func() {
			if res.Header().Get(HeaderContentType) == "application/first" {
				res.Header().Set(HeaderContentType, "application/second")
			}
		})
		_, err = io.WriteString(c.Response(), "<html>hello</html>")
		return err
	})
	server := httptest.NewServer(e)
	defer server.Close()
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		req, err := http.NewRequest(method, server.URL, nil)
		require.NoError(t, err)
		res, err := server.Client().Do(req)
		require.NoError(t, err)
		_, err = io.Copy(io.Discard, res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, "application/second", res.Header.Get(HeaderContentType), method)
	}
}
