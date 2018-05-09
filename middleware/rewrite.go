package middleware

import (
	"regexp"
	"strings"

	"github.com/labstack/echo"
)

type (
	// RewriteConfig defines the config for Rewrite middleware.
	RewriteConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Rules defines the URL path rewrite rules. The values captured in asterisk can be
		// retrieved by index e.g. $1, $2 and so on.
		// Example:
		// "/old":              "/new",
		// "/api/*":            "/$1",
		// "/js/*":             "/public/javascripts/$1",
		// "/users/*/orders/*": "/user/$1/order/$2",
		// Required.
		Rules map[string]string `yaml:"rules"`

		rulesRegex map[*regexp.Regexp]string
	}
)

var (
	// DefaultRewriteConfig is the default Rewrite middleware config.
	DefaultRewriteConfig = RewriteConfig{
		Skipper: DefaultSkipper,
	}
)

// Rewrite returns a Rewrite middleware.
//
// Rewrite middleware rewrites the URL path based on the provided rules.
func Rewrite(rules map[string]string) echo.MiddlewareFunc {
	c := DefaultRewriteConfig
	c.Rules = rules
	return RewriteWithConfig(c)
}

// RewriteWithConfig returns a Rewrite middleware with config.
// See: `Rewrite()`.
func RewriteWithConfig(config RewriteConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Rules == nil {
		panic("echo: rewrite middleware requires url path rewrite rules")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultBodyDumpConfig.Skipper
	}
	config.rulesRegex = map[*regexp.Regexp]string{}

	// Initialize
	for k, v := range config.Rules {
		k = strings.Replace(k, "*", "(\\S*)", -1)
		config.rulesRegex[regexp.MustCompile(k)] = v
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()

			// Rewrite
			for k, v := range config.rulesRegex {
				replacer := captureTokens(k, req.URL.Path)
				if replacer != nil {
					req.URL.Path = replacer.Replace(v)
				}
			}

			return next(c)
		}
	}
}
