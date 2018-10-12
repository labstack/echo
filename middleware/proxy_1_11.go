// +build go1.11

package middleware

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"net/http/httputil"
)

func proxyHTTP(tgt *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(tgt.URL)
	proxy.ErrorHandler = func(resp http.ResponseWriter, req *http.Request, err error) {
		descr := tgt.URL.String()
		if tgt.Name != "" {
			descr = fmt.Sprintf("%s(%s)", tgt.Name, tgt.URL.String())
		}
		c.Logger().Errorf("remote %s unreachable, could not forward: %v", descr, err)
		c.Error(echo.ErrServiceUnavailable)
	}
	proxy.Transport = config.Transport
	return proxy
}
