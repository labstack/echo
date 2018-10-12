// +build go1.11

package middleware

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"net/http/httputil"
)

func proxyHTTP(t *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(t.URL)
	proxy.ErrorHandler = func(resp http.ResponseWriter, req *http.Request, err error) {
		tgt := t.URL.String()
		if t.Name != "" {
			tgt = fmt.Sprintf("%s(%s)", t.Name, t.URL.String())
		}
		c.Logger().Warnf("remote %s unreachable, could not forward: %v", tgt, err)
		c.Error(echo.ErrServiceUnavailable)
	}
	proxy.Transport = config.Transport
	return proxy
}
