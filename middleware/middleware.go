package middleware

import "github.com/labstack/echo"

type (
	// Skipper defines a function to skip middleware. Returning true skips processing
	// the middleware.
	Skipper func(c echo.Context) bool
)

// defaultSkipper returns false which processes the middleware.
func defaultSkipper(c echo.Context) bool {
	return false
}
