// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestRouterPathParamsDecodeOnce(t *testing.T) {
	values := []struct {
		encoded, decoded string
		hasRawPath       bool
	}{
		{encoded: "plain", decoded: "plain"},
		{encoded: "100%2525", decoded: "100%25"},
		{encoded: "a%2520b", decoded: "a%20b"},
		{encoded: "%252e%252e", decoded: "%2e%2e"},
		{encoded: "a%252Fb", decoded: "a%2Fb"},
		{encoded: "a%20b", decoded: "a b"},
		{encoded: "a+b", decoded: "a+b"},
		{encoded: "%41", decoded: "A", hasRawPath: true},
		{encoded: "a%2Fb", decoded: "a/b", hasRawPath: true},
		{encoded: "a%2525%2Fb", decoded: "a%25/b", hasRawPath: true},
	}
	for _, concurrent := range []bool{false, true} {
		for _, decodedMatching := range []bool{false, true} {
			for _, unescape := range []bool{false, true} {
				for _, wildcard := range []bool{false, true} {
					name := fmt.Sprintf("concurrent=%t/decoded=%t/unescape=%t/wildcard=%t", concurrent, decodedMatching, unescape, wildcard)
					t.Run(name, func(t *testing.T) {
						var router echo.Router = echo.NewRouter(echo.RouterConfig{
							UseEscapedPathForMatching: decodedMatching,
							UnescapePathParamValues:   unescape,
						})
						if concurrent {
							router = echo.NewConcurrentRouter(router)
						}
						e := echo.NewWithConfig(echo.Config{Router: router})
						route, param := "/items/:id", "id"
						if wildcard {
							route, param = "/items/*", "*"
						}
						e.GET(route, func(c *echo.Context) error {
							return c.String(http.StatusOK, c.Param(param))
						})
						for _, tc := range values {
							t.Run(tc.encoded, func(t *testing.T) {
								req := httptest.NewRequest(http.MethodGet, "/items/"+tc.encoded, nil)
								rec := httptest.NewRecorder()
								e.ServeHTTP(rec, req)
								want := tc.decoded
								if !decodedMatching && !unescape && tc.hasRawPath {
									want = tc.encoded
								}
								if rec.Code != http.StatusOK || rec.Body.String() != want {
									t.Errorf("status=%d param=%q, want 200 and %q", rec.Code, rec.Body.String(), want)
								}
							})
						}
					})
				}
			}
		}
	}
}
