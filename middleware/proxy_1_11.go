// +build go1.11

package middleware

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/labstack/echo"
)

func proxyHTTP(tgt *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(tgt.URL)
	if c.Request().Header.Get(echo.HeaderAccept) == "text/event-stream" {
		proxy.FlushInterval = 100 * time.Millisecond
	}
	proxy.ErrorHandler = func(resp http.ResponseWriter, req *http.Request, err error) {
		desc := tgt.URL.String()
		if tgt.Name != "" {
			desc = fmt.Sprintf("%s(%s)", tgt.Name, tgt.URL.String())
		}
		c.Logger().Errorf("remote %s unreachable, could not forward: %v", desc, err)
		c.Error(echo.NewHTTPError(http.StatusServiceUnavailable))
	}
	proxy.Transport = config.Transport
	return proxy
}
