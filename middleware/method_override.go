package middleware

import (
	"github.com/labstack/echo"
)

const (
	HttpMethodOverrideHeader = "X-HTTP-Method-Override"
)

func OverrideMethod() echo.MiddlewareFunc {
	return Override()
}

// Override checks for the X-HTTP-Method-Override header
// or the body for parameter, `_method`
// and uses the http method instead of Request.Method.
// It isn't secure to override e.g a GET to a POST,
// so only Request.Method which are POSTs are considered.
func Override() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			originalMethod := c.Request().Method()

			if originalMethod == "POST" {
				m := c.FormValue("_method")
				if m != "" {
					c.Request().SetMethod(m)
				}
				m = c.Request().Header().Get(HttpMethodOverrideHeader)
				if m != "" {
					c.Request().SetMethod(m)
				}
			}

			return next(c)
		}
	}
}
