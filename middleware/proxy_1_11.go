// +build go1.11

package middleware

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/labstack/echo/v4"
)

func proxyHTTP(tgt *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(tgt.URL)
	proxy.ErrorHandler = func(resp http.ResponseWriter, req *http.Request, err error) {
		desc := tgt.URL.String()
		if tgt.Name != "" {
			desc = fmt.Sprintf("%s(%s)", tgt.Name, tgt.URL.String())
		}
		httpError := echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("remote %s unreachable, could not forward: %v", desc, err))
		httpError.Internal = err
		c.Set("_error", httpError)
	}
	proxy.Transport = config.Transport
	proxy.ModifyResponse = config.ModifyResponse
	return proxy
}
