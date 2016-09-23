package middleware

import (
	"fmt"

	"github.com/labstack/echo"
)

type (
	// SecureConfig defines the config for Secure middleware.
	SecureConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// XSSProtection provides protection against cross-site scripting attack (XSS)
		// by setting the `X-XSS-Protection` header.
		// Optional. Default value "1; mode=block".
		XSSProtection string `json:"xss_protection"`

		// ContentTypeNosniff provides protection against overriding Content-Type
		// header by setting the `X-Content-Type-Options` header.
		// Optional. Default value "nosniff".
		ContentTypeNosniff string `json:"content_type_nosniff"`

		// XFrameOptions can be used to indicate whether or not a browser should
		// be allowed to render a page in a <frame>, <iframe> or <object> .
		// Sites can use this to avoid clickjacking attacks, by ensuring that their
		// content is not embedded into other sites.provides protection against
		// clickjacking.
		// Optional. Default value "SAMEORIGIN".
		// Possible values:
		// - "SAMEORIGIN" - The page can only be displayed in a frame on the same origin as the page itself.
		// - "DENY" - The page cannot be displayed in a frame, regardless of the site attempting to do so.
		// - "ALLOW-FROM uri" - The page can only be displayed in a frame on the specified origin.
		XFrameOptions string `json:"x_frame_options"`

		// HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how
		// long (in seconds) browsers should remember that this site is only to
		// be accessed using HTTPS. This reduces your exposure to some SSL-stripping
		// man-in-the-middle (MITM) attacks.
		// Optional. Default value 0.
		HSTSMaxAge int `json:"hsts_max_age"`

		// HSTSExcludeSubdomains won't include subdomains tag in the `Strict Transport Security`
		// header, excluding all subdomains from security policy. It has no effect
		// unless HSTSMaxAge is set to a non-zero value.
		// Optional. Default value false.
		HSTSExcludeSubdomains bool `json:"hsts_exclude_subdomains"`

		// ContentSecurityPolicy sets the `Content-Security-Policy` header providing
		// security against cross-site scripting (XSS), clickjacking and other code
		// injection attacks resulting from execution of malicious content in the
		// trusted web page context.
		// Optional. Default value "".
		ContentSecurityPolicy string `json:"content_security_policy"`
	}
)

var (
	// DefaultSecureConfig is the default Secure middleware config.
	DefaultSecureConfig = SecureConfig{
		Skipper:            defaultSkipper,
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "SAMEORIGIN",
	}
)

// Secure returns a Secure middleware.
// Secure middleware provides protection against cross-site scripting (XSS) attack,
// content type sniffing, clickjacking, insecure connection and other code injection
// attacks.
func Secure() echo.MiddlewareFunc {
	return SecureWithConfig(DefaultSecureConfig)
}

// SecureWithConfig returns a Secure middleware with config.
// See: `Secure()`.
func SecureWithConfig(config SecureConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultSecureConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			if config.XSSProtection != "" {
				res.Header().Set(echo.HeaderXXSSProtection, config.XSSProtection)
			}
			if config.ContentTypeNosniff != "" {
				res.Header().Set(echo.HeaderXContentTypeOptions, config.ContentTypeNosniff)
			}
			if config.XFrameOptions != "" {
				res.Header().Set(echo.HeaderXFrameOptions, config.XFrameOptions)
			}
			if (c.IsTLS() || (req.Header.Get(echo.HeaderXForwardedProto) == "https")) && config.HSTSMaxAge != 0 {
				subdomains := ""
				if !config.HSTSExcludeSubdomains {
					subdomains = "; includeSubdomains"
				}
				res.Header().Set(echo.HeaderStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", config.HSTSMaxAge, subdomains))
			}
			if config.ContentSecurityPolicy != "" {
				res.Header().Set(echo.HeaderContentSecurityPolicy, config.ContentSecurityPolicy)
			}
			return next(c)
		}
	}
}
