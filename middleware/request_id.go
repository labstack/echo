package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
)

type (
	// RequestIDConfig defines the config for RequestID middleware.
	RequestIDConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		//Define the maximun length for RequestID value
		MaxLength uint8

		// Generator defines a function to generate an ID.
		// Optional. Default value random.String(32).
		Generator func() string

		// RequestIDHandler defines a function which is executed for a request id.
		RequestIDHandler func(echo.Context, string)
	}
)

var (
	// DefaultRequestIDConfig is the default RequestID middleware config.
	DefaultRequestIDConfig = RequestIDConfig{
		Skipper:   DefaultSkipper,
		Generator: generator,
	}
	ErrRequestIDMaxLength = echo.NewHTTPError(http.StatusNotAcceptable, "request id length should not be greater than MaxLength")
	RequestIDLength       = uint8(32)
)

// RequestID returns a X-Request-ID middleware.
func RequestID() echo.MiddlewareFunc {
	return RequestIDWithConfig(DefaultRequestIDConfig)
}

// RequestIDWithConfig returns a X-Request-ID middleware with config.
func RequestIDWithConfig(config RequestIDConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultRequestIDConfig.Skipper
	}
	if config.Generator == nil {
		config.Generator = generator
	}
	if config.MaxLength == 0 {
		config.MaxLength = RequestIDLength
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			rid := req.Header.Get(echo.HeaderXRequestID)

			if len(rid) > int(config.MaxLength) {
				c.Error(&echo.HTTPError{
					Code:    ErrRequestIDMaxLength.Code,
					Message: ErrRequestIDMaxLength.Message,
				})
				return nil
			}

			if rid == "" {
				rid = config.Generator()
			}
			res.Header().Set(echo.HeaderXRequestID, rid)
			if config.RequestIDHandler != nil {
				config.RequestIDHandler(c, rid)
			}

			return next(c)
		}
	}
}

func generator() string {
	return random.String(RequestIDLength)
}
