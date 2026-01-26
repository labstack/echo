// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
)

// CORSConfig defines the config for CORS middleware.
type CORSConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// AllowOrigins determines the value of the Access-Control-Allow-Origin
	// response header.  This header defines a list of origins that may access the
	// resource.
	//
	// Origin consist of following parts: `scheme + "://" + host + optional ":" + port`
	// Wildcard can be used, but has to be set explicitly []string{"*"}
	// Example: `https://example.com`, `http://example.com:8080`, `*`
	//
	// Security: use extreme caution when handling the origin, and carefully
	// validate any logic. Remember that attackers may register hostile domain names.
	// See https://blog.portswigger.net/2016/10/exploiting-cors-misconfigurations-for.html
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	//
	// Mandatory.
	AllowOrigins []string

	// UnsafeAllowOriginFunc is an optional custom function to validate the origin. It takes the
	// origin as an argument and returns
	// - string, allowed origin
	// - bool, true if allowed or false otherwise.
	// - error, if an error is returned, it is returned immediately by the handler.
	// If this option is set, AllowOrigins is ignored.
	//
	// Security: use extreme caution when handling the origin, and carefully
	// validate any logic. Remember that attackers may register hostile (sub)domain names.
	// See https://blog.portswigger.net/2016/10/exploiting-cors-misconfigurations-for.html
	//
	// Sub-domain checks example:
	// 		UnsafeAllowOriginFunc: func(c *echo.Context, origin string) (string, bool, error) {
	//			if strings.HasSuffix(origin, ".example.com") {
	//				return origin, true, nil
	//			}
	//			return "", false, nil
	//		},
	//
	// Optional.
	UnsafeAllowOriginFunc func(c *echo.Context, origin string) (allowedOrigin string, allowed bool, err error)

	// AllowMethods determines the value of the Access-Control-Allow-Methods
	// response header.  This header specified the list of methods allowed when
	// accessing the resource.  This is used in response to a preflight request.
	//
	// Optional. Default value DefaultCORSConfig.AllowMethods.
	// If `allowMethods` is left empty, this middleware will fill for preflight
	// request `Access-Control-Allow-Methods` header value
	// from `Allow` header that echo.Router set into context.
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
	AllowMethods []string

	// AllowHeaders determines the value of the Access-Control-Allow-Headers
	// response header.  This header is used in response to a preflight request to
	// indicate which HTTP headers can be used when making the actual request.
	//
	// Optional. Defaults to empty list. No domains allowed for CORS.
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
	AllowHeaders []string

	// AllowCredentials determines the value of the
	// Access-Control-Allow-Credentials response header.  This header indicates
	// whether or not the response to the request can be exposed when the
	// credentials mode (Request.credentials) is true. When used as part of a
	// response to a preflight request, this indicates whether or not the actual
	// request can be made using credentials.  See also
	// [MDN: Access-Control-Allow-Credentials].
	//
	// Optional. Default value false, in which case the header is not set.
	//
	// Security: avoid using `AllowCredentials = true` with `AllowOrigins = *`.
	// See "Exploiting CORS misconfigurations for Bitcoins and bounties",
	// https://blog.portswigger.net/2016/10/exploiting-cors-misconfigurations-for.html
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
	AllowCredentials bool

	// ExposeHeaders determines the value of Access-Control-Expose-Headers, which
	// defines a list of headers that clients are allowed to access.
	//
	// Optional. Default value []string{}, in which case the header is not set.
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Header
	ExposeHeaders []string

	// MaxAge determines the value of the Access-Control-Max-Age response header.
	// This header indicates how long (in seconds) the results of a preflight
	// request can be cached.
	// The header is set only if MaxAge != 0, negative value sends "0" which instructs browsers not to cache that response.
	//
	// Optional. Default value 0 - meaning header is not sent.
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age
	MaxAge int
}

// CORS returns a Cross-Origin Resource Sharing (CORS) middleware.
// See also [MDN: Cross-Origin Resource Sharing (CORS)].
//
// Origin consist of following parts: `scheme + "://" + host + optional ":" + port`
// Wildcard `*` can be used, but has to be set explicitly.
// Example: `https://example.com`, `http://example.com:8080`, `*`
//
// Security: Poorly configured CORS can compromise security because it allows
// relaxation of the browser's Same-Origin policy.  See [Exploiting CORS
// misconfigurations for Bitcoins and bounties] and [Portswigger: Cross-origin
// resource sharing (CORS)] for more details.
//
// [MDN: Cross-Origin Resource Sharing (CORS)]: https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS
// [Exploiting CORS misconfigurations for Bitcoins and bounties]: https://blog.portswigger.net/2016/10/exploiting-cors-misconfigurations-for.html
// [Portswigger: Cross-origin resource sharing (CORS)]: https://portswigger.net/web-security/cors
func CORS(allowOrigins ...string) echo.MiddlewareFunc {
	c := CORSConfig{
		AllowOrigins: allowOrigins,
	}
	return CORSWithConfig(c)
}

// CORSWithConfig returns a CORS middleware with config or panics on invalid configuration.
// See: [CORS].
func CORSWithConfig(config CORSConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts CORSConfig to middleware or returns an error for invalid configuration
func (config CORSConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}
	hasCustomAllowMethods := true
	if len(config.AllowMethods) == 0 {
		hasCustomAllowMethods = false
		config.AllowMethods = []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete}
	}

	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")

	maxAge := "0"
	if config.MaxAge > 0 {
		maxAge = strconv.Itoa(config.MaxAge)
	}

	allowOriginFunc := config.UnsafeAllowOriginFunc
	if config.UnsafeAllowOriginFunc == nil {
		if len(config.AllowOrigins) == 0 {
			return nil, errors.New("at least one AllowOrigins is required or UnsafeAllowOriginFunc must be provided")
		}
		allowOriginFunc = config.defaultAllowOriginFunc
		for _, origin := range config.AllowOrigins {
			if origin == "*" {
				if config.AllowCredentials {
					return nil, fmt.Errorf("* as allowed origin and AllowCredentials=true is insecure and not allowed. Use custom UnsafeAllowOriginFunc")
				}
				allowOriginFunc = config.starAllowOriginFunc
				break
			}
			if err := validateOrigin(origin, "allow origin"); err != nil {
				return nil, err
			}
		}
		config.AllowOrigins = append([]string(nil), config.AllowOrigins...)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			origin := req.Header.Get(echo.HeaderOrigin)

			res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)

			// Preflight request is an OPTIONS request, using three HTTP request headers: Access-Control-Request-Method,
			// Access-Control-Request-Headers, and the Origin header. See: https://developer.mozilla.org/en-US/docs/Glossary/Preflight_request
			// For simplicity we just consider method type and later `Origin` header.
			preflight := req.Method == http.MethodOptions

			// Although router adds special handler in case of OPTIONS method we avoid calling next for OPTIONS in this middleware
			// as CORS requests do not have cookies / authentication headers by default, so we could get stuck in auth
			// middlewares by calling next(c).
			// But we still want to send `Allow` header as response in case of Non-CORS OPTIONS request as router default
			// handler does.
			routerAllowMethods := ""
			if preflight {
				tmpAllowMethods, ok := c.Get(echo.ContextKeyHeaderAllow).(string)
				if ok && tmpAllowMethods != "" {
					routerAllowMethods = tmpAllowMethods
					c.Response().Header().Set(echo.HeaderAllow, routerAllowMethods)
				}
			}

			// No Origin provided. This is (probably) not request from actual browser - proceed executing middleware chain
			if origin == "" {
				if preflight { // req.Method=OPTIONS
					return c.NoContent(http.StatusNoContent)
				}
				return next(c) // let non-browser calls through
			}

			allowedOrigin, allowed, err := allowOriginFunc(c, origin)
			if err != nil {
				return err
			}
			if !allowed {
				// Origin existed and was NOT allowed
				if preflight {
					// From: https://github.com/labstack/echo/issues/2767
					// If the request's origin isn't allowed by the CORS configuration,
					// the middleware should simply omit the relevant CORS headers from the response
					// and let the browser fail the CORS check (if any).
					return c.NoContent(http.StatusNoContent)
				}
				// From: https://github.com/labstack/echo/issues/2767
				// no CORS middleware should block non-preflight requests;
				// such requests should be let through. One reason is that not all requests that
				// carry an Origin header participate in the CORS protocol.
				return next(c)
			}

			// Origin existed and was allowed

			res.Header().Set(echo.HeaderAccessControlAllowOrigin, allowedOrigin)
			if config.AllowCredentials {
				res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
			}

			// Simple request will be let though
			if !preflight {
				if exposeHeaders != "" {
					res.Header().Set(echo.HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				return next(c)
			}
			// Below code is for Preflight (OPTIONS) request
			//
			// Preflight will end with c.NoContent(http.StatusNoContent) as we do not know if
			// at the end of handler chain is actual OPTIONS route or 404/405 route which
			// response code will confuse browsers
			res.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestMethod)
			res.Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestHeaders)

			if !hasCustomAllowMethods && routerAllowMethods != "" {
				res.Header().Set(echo.HeaderAccessControlAllowMethods, routerAllowMethods)
			} else {
				res.Header().Set(echo.HeaderAccessControlAllowMethods, allowMethods)
			}

			if allowHeaders != "" {
				res.Header().Set(echo.HeaderAccessControlAllowHeaders, allowHeaders)
			} else {
				h := req.Header.Get(echo.HeaderAccessControlRequestHeaders)
				if h != "" {
					res.Header().Set(echo.HeaderAccessControlAllowHeaders, h)
				}
			}
			if config.MaxAge != 0 {
				res.Header().Set(echo.HeaderAccessControlMaxAge, maxAge)
			}
			return c.NoContent(http.StatusNoContent)
		}
	}, nil
}

func (config CORSConfig) starAllowOriginFunc(c *echo.Context, origin string) (string, bool, error) {
	return "*", true, nil
}

func (config CORSConfig) defaultAllowOriginFunc(c *echo.Context, origin string) (string, bool, error) {
	for _, allowedOrigin := range config.AllowOrigins {
		if strings.EqualFold(allowedOrigin, origin) {
			return allowedOrigin, true, nil
		}
	}
	return "", false, nil
}
