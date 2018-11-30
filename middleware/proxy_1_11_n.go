// +build !go1.11

package middleware

import (
	"github.com/labstack/echo"
	"net/http"
	"net/http/httputil"
)

func proxyHTTP(t *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	return httputil.NewSingleHostReverseProxy(t.URL)
}
