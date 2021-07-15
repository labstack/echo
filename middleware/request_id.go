package middleware

import (
	"github.com/labstack/echo/v5"
)

// RequestIDConfig defines the config for RequestID middleware.
type RequestIDConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Generator defines a function to generate an ID.
	// Optional. Default value random.String(32).
	Generator func() string

	// RequestIDHandler defines a function which is executed for a request id.
	RequestIDHandler func(c echo.Context, requestID string)

	// TargetHeader defines what header to look for to populate the id
	TargetHeader string
}

// RequestID returns a X-Request-ID middleware.
func RequestID() echo.MiddlewareFunc {
	return RequestIDWithConfig(RequestIDConfig{})
}

// RequestIDWithConfig returns a X-Request-ID middleware with config or panics on invalid configuration.
func RequestIDWithConfig(config RequestIDConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts RequestIDConfig to middleware or returns an error for invalid configuration
func (config RequestIDConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	if config.Generator == nil {
		config.Generator = createRandomStringGenerator(32)
	}
	if config.TargetHeader == "" {
		config.TargetHeader = echo.HeaderXRequestID
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			rid := req.Header.Get(config.TargetHeader)
			if rid == "" {
				rid = config.Generator()
			}
			res.Header().Set(config.TargetHeader, rid)
			if config.RequestIDHandler != nil {
				config.RequestIDHandler(c, rid)
			}

			return next(c)
		}
	}, nil
}
