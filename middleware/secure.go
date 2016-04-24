package middleware

import (
	"fmt"

	"github.com/labstack/echo"
)

type (
	SecureConfig struct {
		STSMaxAge             int64
		STSIncludeSubdomains  bool
		FrameDeny             bool
		FrameOptionsValue     string
		ContentTypeNosniff    bool
		XssProtected          bool
		XssProtectionValue    string
		ContentSecurityPolicy string
		DisableProdCheck      bool
	}
)

var (
	DefaultSecureConfig = SecureConfig{}
)

const (
	stsHeader           = "Strict-Transport-Security"
	stsSubdomainString  = "; includeSubdomains"
	frameOptionsHeader  = "X-Frame-Options"
	frameOptionsValue   = "DENY"
	contentTypeHeader   = "X-Content-Type-Options"
	contentTypeValue    = "nosniff"
	xssProtectionHeader = "X-XSS-Protection"
	xssProtectionValue  = "1; mode=block"
	cspHeader           = "Content-Security-Policy"
)

func Secure() echo.MiddlewareFunc {
	return SecureWithConfig(DefaultSecureConfig)
}

func SecureWithConfig(config SecureConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			setFrameOptions(c, config)
			setContentTypeOptions(c, config)
			setXssProtection(c, config)
			setSTS(c, config)
			setCSP(c, config)
			return next(c)
		}
	}
}

func setFrameOptions(c echo.Context, opts SecureConfig) {
	if opts.FrameOptionsValue != "" {
		c.Response().Header().Set(frameOptionsHeader, opts.FrameOptionsValue)
	} else if opts.FrameDeny {
		c.Response().Header().Set(frameOptionsHeader, frameOptionsValue)
	}
}

func setContentTypeOptions(c echo.Context, opts SecureConfig) {
	if opts.ContentTypeNosniff {
		c.Response().Header().Set(contentTypeHeader, contentTypeValue)
	}
}

func setXssProtection(c echo.Context, opts SecureConfig) {
	if opts.XssProtectionValue != "" {
		c.Response().Header().Set(xssProtectionHeader, opts.XssProtectionValue)
	} else if opts.XssProtected {
		c.Response().Header().Set(xssProtectionHeader, xssProtectionValue)
	}
}

func setSTS(c echo.Context, opts SecureConfig) {
	if opts.STSMaxAge != 0 && opts.DisableProdCheck {
		subDomains := ""
		if opts.STSIncludeSubdomains {
			subDomains = stsSubdomainString
		}

		c.Response().Header().Set(stsHeader, fmt.Sprintf("max-age=%d%s", opts.STSMaxAge, subDomains))
	}
}

func setCSP(c echo.Context, opts SecureConfig) {
	if opts.ContentSecurityPolicy != "" {
		c.Response().Header().Set(cspHeader, opts.ContentSecurityPolicy)
	}
}
