package middleware

import "github.com/labstack/echo"

type (
	// Skipper defines a function to skip middleware. Returning true skips processing
	// the middleware.
	Skipper func(c echo.Context) bool
)

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(c echo.Context) bool {
	return false
}
