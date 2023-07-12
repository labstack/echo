package middleware

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

var (
	protoRegex = regexp.MustCompile(`(?i)(?:proto=)(https|http)`)
	ipRegex    = regexp.MustCompile("(?i)(?:for=)([^(;|,| )]+)")
)

func ProxyHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if fwd := c.Request().Header.Get(echo.HeaderForwarded); fwd != "" {
				if match := ipRegex.FindStringSubmatch(fwd); len(match) > 1 {
					c.Request().RemoteAddr = strings.Trim(match[1], `"`)
				}
			} else if fwd := c.RealIP(); fwd != "" {
				c.Request().RemoteAddr = fwd
			}

			if scheme := getScheme(c.Request()); scheme != "" {
				c.Request().URL.Scheme = scheme
			}

			if c.Request().Header.Get(echo.HeaderXForwardedHost) != "" {
				c.Request().Host = c.Request().Header.Get(echo.HeaderXForwardedHost)
			}

			if prefix := c.Request().Header.Get(echo.HeaderXForwardedPrefix); prefix != "" {
				c.Request().RequestURI, _ = url.JoinPath(prefix, c.Request().RequestURI)
				c.Request().URL.Path, _ = url.JoinPath(prefix, c.Request().URL.Path)
			}
			return next(c)
		}
	}
}

func getScheme(r *http.Request) string {
	var scheme string

	if proto := r.Header.Get(echo.HeaderXForwardedProto); proto != "" {
		scheme = strings.ToLower(proto)
	} else if proto := r.Header.Get(echo.HeaderXForwardedProtocol); proto != "" {
		scheme = strings.ToLower(proto)
	} else if proto = r.Header.Get(echo.HeaderForwarded); proto != "" {
		if match := protoRegex.FindStringSubmatch(proto); len(match) > 1 {
			scheme = strings.ToLower(match[1])
		}
	}
	return scheme
}
