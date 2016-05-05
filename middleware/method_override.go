package middleware

import (
	"github.com/labstack/echo"
)

type (
	// MethodOverrideConfig defines the config for method override middleware.
	MethodOverrideConfig struct {
		// Getter is a function that gets overridden method from the request.
		Getter MethodOverrideGetter
	}

	// MethodOverrideGetter is a function that gets overridden method from the request
	// Optional, with default values as `MethodFromHeader(echo.HeaderXHTTPMethodOverride)`.
	MethodOverrideGetter func(echo.Context) string
)

var (
	// DefaultMethodOverrideConfig is the default method override middleware config.
	DefaultMethodOverrideConfig = MethodOverrideConfig{
		Getter: MethodFromHeader(echo.HeaderXHTTPMethodOverride),
	}
)

// MethodOverride returns a method override middleware.
// MethodOverride  middleware checks for the overridden method from the request and
// uses it instead of the original method.
//
// For security reasons, only `POST` method can be overridden.
func MethodOverride() echo.MiddlewareFunc {
	return MethodOverrideWithConfig(DefaultMethodOverrideConfig)
}

// MethodOverrideWithConfig returns a method override middleware from config.
// See `MethodOverride()`.
func MethodOverrideWithConfig(config MethodOverrideConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Getter == nil {
		config.Getter = DefaultMethodOverrideConfig.Getter
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			if req.Method() == echo.POST {
				m := config.Getter(c)
				if m != "" {
					req.SetMethod(m)
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
		return c.Request().Header().Get(header)
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
