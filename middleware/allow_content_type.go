package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// AllowContentType returns an AllowContentType middleware
//
// It requries at least one content type to be passed in as an argument
func AllowContentType(contentTypes ...string) echo.MiddlewareFunc {
	if len(contentTypes) == 0 {
		panic("echo: allow-content middleware requires at least one content type")
	}
	allowedContentTypes := make(map[string]struct{}, len(contentTypes))
	for _, ctype := range contentTypes {
		allowedContentTypes[strings.TrimSpace(strings.ToLower(ctype))] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().ContentLength == 0 {
				// skip check for empty content body
				return next(c)
			}
			s := strings.ToLower(strings.TrimSpace(c.Request().Header.Get("Content-Type")))
			if i := strings.Index(s, ";"); i > -1 {
				s = s[0:i]
			}
			if _, ok := allowedContentTypes[s]; ok {
				return next(c)
			}
			return echo.NewHTTPError(http.StatusUnsupportedMediaType)
		}
	}
}
