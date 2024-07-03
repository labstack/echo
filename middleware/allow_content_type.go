package middleware

import (
	"mime"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
)

func AllowContentType(contentTypes ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			mediaType, _, err := mime.ParseMediaType(c.Request().Header.Get("Content-Type"))
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid content-type value")
			}
			if slices.Contains(contentTypes, mediaType) {
				return next(c)
			}
			return echo.NewHTTPError(http.StatusUnsupportedMediaType)
		}
	}
}
