// +build go1.11

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/labstack/echo/v4"
)

// StatusCodeContextCanceled is a custom HTTP status code for situations
// where a client unexpectedly closed the connection to the server.
// As there is no standard error code for "client closed connection", but
// various well-known HTTP clients and server implement this HTTP code we use
// 499 too instead of the more problematic 5xx, which does not allow to detect this situation
const StatusCodeContextCanceled = 499

func proxyHTTP(tgt *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(tgt.URL)
	proxy.ErrorHandler = func(resp http.ResponseWriter, req *http.Request, err error) {
		desc := tgt.URL.String()
		if tgt.Name != "" {
			desc = fmt.Sprintf("%s(%s)", tgt.Name, tgt.URL.String())
		}
		// If the client canceled the request (usually by closing the connection), we can report a
		// client error (4xx) instead of a server error (5xx) to correctly identify the situation.
		// The Go standard library (at of late 2020) wraps the exported, standard
		// context.Canceled error with unexported garbage value requiring a substring check, see
		// https://github.com/golang/go/blob/6965b01ea248cabb70c3749fd218b36089a21efb/src/net/net.go#L416-L430
		if err == context.Canceled || strings.Contains(err.Error(), "operation was canceled") {
			httpError := echo.NewHTTPError(StatusCodeContextCanceled, fmt.Sprintf("client closed connection: %v", err))
			httpError.Internal = err
			c.Set("_error", httpError)
		} else {
			httpError := echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("remote %s unreachable, could not forward: %v", desc, err))
			httpError.Internal = err
			c.Set("_error", httpError)
		}
	}
	proxy.Transport = config.Transport
	proxy.ModifyResponse = config.ModifyResponse
	return proxy
}
