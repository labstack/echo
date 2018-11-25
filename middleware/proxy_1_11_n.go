// +build !go1.11

package middleware

import (
	"net/http"
	"net/http/httputil"

	"github.com/labstack/echo"
)

func proxyHTTP(t *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(t.URL)
	proxy.Transport = config.Transport
	return proxy
}
