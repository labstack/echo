// +build !go1.11

package middleware

import (
	"net/http"
	"net/http/httputil"

	"github.com/labstack/echo/v4"
)

func proxyHTTP(t *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	return httputil.NewSingleHostReverseProxy(t.URL)
}
