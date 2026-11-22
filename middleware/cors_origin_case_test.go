// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestCORSReflectsMatchedOrigin(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodOptions} {
		t.Run(method, func(t *testing.T) {
			e := echo.New()
			e.Use(CORSWithConfig(CORSConfig{
				AllowOrigins:     []string{"https://EXAMPLE.com"},
				AllowCredentials: true,
			}))
			e.GET("/", func(c *echo.Context) error {
				return c.String(http.StatusOK, "allowed")
			})
			req := httptest.NewRequest(method, "/", nil)
			req.Header.Set(echo.HeaderOrigin, "https://example.com")
			if method == http.MethodOptions {
				req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, req.Header.Get(echo.HeaderOrigin), rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
			assert.Equal(t, "true", rec.Header().Get(echo.HeaderAccessControlAllowCredentials))
		})
	}
}
