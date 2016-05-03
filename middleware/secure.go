package middleware

import (
	"fmt"

	"github.com/labstack/echo"
)

type (
	SecureConfig struct {
		DisableXSSProtection         bool
		DisableContentTypeNosniff    bool
		XFrameOptions                string
		DisableHSTSIncludeSubdomains bool
		HSTSMaxAge                   int
		ContentSecurityPolicy        string
	}
)

var (
	DefaultSecureConfig = SecureConfig{
		XFrameOptions: "SAMEORIGIN",
	}
)

func Secure() echo.MiddlewareFunc {
	return SecureWithConfig(DefaultSecureConfig)
}

func SecureWithConfig(config SecureConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !config.DisableXSSProtection {
				c.Response().Header().Set(echo.HeaderXXSSProtection, "1; mode=block")
			}
			if !config.DisableContentTypeNosniff {
				c.Response().Header().Set(echo.HeaderXContentTypeOptions, "nosniff")
			}
			if config.XFrameOptions != "" {
				c.Response().Header().Set(echo.HeaderXFrameOptions, config.XFrameOptions)
			}
			if config.HSTSMaxAge != 0 {
				subdomains := ""
				if !config.DisableHSTSIncludeSubdomains {
					subdomains = "; includeSubdomains"
				}
				c.Response().Header().Set(echo.HeaderStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", config.HSTSMaxAge, subdomains))
			}
			if config.ContentSecurityPolicy != "" {
				c.Response().Header().Set(echo.HeaderContentSecurityPolicy, config.ContentSecurityPolicy)
			}
			return next(c)
		}
	}
}
