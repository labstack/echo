// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// MethodOverrideConfig defines the config for MethodOverride middleware.
type MethodOverrideConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Getter is a function that gets overridden method from the request.
	// Optional. Default values MethodFromHeader(echo.HeaderXHTTPMethodOverride).
	Getter MethodOverrideGetter
}

// MethodOverrideGetter is a function that gets overridden method from the request
type MethodOverrideGetter func(echo.Context) string

// DefaultMethodOverrideConfig is the default MethodOverride middleware config.
var DefaultMethodOverrideConfig = MethodOverrideConfig{
	Skipper: DefaultSkipper,
	Getter:  MethodFromHeader(echo.HeaderXHTTPMethodOverride),
}

// MethodOverride returns a MethodOverride middleware.
// MethodOverride  middleware checks for the overridden method from the request and
// uses it instead of the original method.
//
// For security reasons, only `POST` method can be overridden, and it cannot be overridden to `GET`, `HEAD`,
// `OPTIONS`, `TRACE` or `CONNECT`. Otherwise a cross-site form POST could skip checks that apply only to
// state-changing methods, such as the CSRF middleware.
//
// Register it with Echo#Pre so that routing uses the overridden method.
func MethodOverride() echo.MiddlewareFunc {
	return MethodOverrideWithConfig(DefaultMethodOverrideConfig)
}

// MethodOverrideWithConfig returns a MethodOverride middleware with config.
// See: `MethodOverride()`.
func MethodOverrideWithConfig(config MethodOverrideConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultMethodOverrideConfig.Skipper
	}
	if config.Getter == nil {
		config.Getter = DefaultMethodOverrideConfig.Getter
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			if req.Method == http.MethodPost {
				m := config.Getter(c)
				if m != "" && !isForbiddenOverrideMethod(m) {
					req.Method = m
				}
			}
			return next(c)
		}
	}
}

// MethodFromHeader is a `MethodOverrideGetter` that gets overridden method from
// the request header.
func MethodFromHeader(header string) MethodOverrideGetter {
	return func(c echo.Context) string {
		return c.Request().Header.Get(header)
	}
}

// MethodFromForm is a `MethodOverrideGetter` that gets overridden method from the
// form parameter.
func MethodFromForm(param string) MethodOverrideGetter {
	return func(c echo.Context) string {
		return c.FormValue(param)
	}
}

// MethodFromQuery is a `MethodOverrideGetter` that gets overridden method from
// the query parameter.
func MethodFromQuery(param string) MethodOverrideGetter {
	return func(c echo.Context) string {
		return c.QueryParam(param)
	}
}

// isForbiddenOverrideMethod reports whether POST must not be overridden to method m. Safe methods (and CONNECT) are
// forbidden because middlewares such as CSRF do not check them.
func isForbiddenOverrideMethod(m string) bool {
	for _, forbidden := range []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace, http.MethodConnect} {
		if strings.EqualFold(m, forbidden) {
			return true
		}
	}
	return false
}
