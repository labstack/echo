// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// Skipper defines a function to skip middleware. Returning true skips processing
// the middleware.
type Skipper func(c echo.Context) bool

// BeforeFunc defines a function which is executed just before the middleware.
type BeforeFunc func(c echo.Context)

func captureTokens(pattern *regexp.Regexp, input string) *strings.Replacer {
	groups := pattern.FindAllStringSubmatch(input, -1)
	if groups == nil {
		return nil
	}
	values := groups[0][1:]
	replace := make([]string, 2*len(values))
	for i, v := range values {
		j := 2 * i
		replace[j] = "$" + strconv.Itoa(i+1)
		replace[j+1] = v
	}
	return strings.NewReplacer(replace...)
}

func rewriteRulesRegex(rewrite map[string]string) map[*regexp.Regexp]string {
	// Initialize
	rulesRegex := map[*regexp.Regexp]string{}
	for k, v := range rewrite {
		k = regexp.QuoteMeta(k)
		k = strings.ReplaceAll(k, `\*`, "(.*?)")
		if strings.HasPrefix(k, `\^`) {
			k = strings.ReplaceAll(k, `\^`, "^")
		}
		k = k + "$"
		rulesRegex[regexp.MustCompile(k)] = v
	}
	return rulesRegex
}

// rewriteURL rewrites req.URL with the first matching rule.
//
// With routed set (the router has already chosen a route), the target is applied as a literal path and query
// instead of being resolved with req.URL.Parse, and a target with a "." or ".." segment or a leading "//" is rejected.
// Resolving those after routing would move the request to a path that the router never matched, for example from an
// unprotected /api/* route to a protected /api/admin group (GHSA-w5j2-g3fc-pwr6).
func rewriteURL(rewriteRegex map[*regexp.Regexp]string, req *http.Request, routed bool) error {
	if len(rewriteRegex) == 0 {
		return nil
	}

	// Depending on how HTTP request is sent RequestURI could contain Scheme://Host/path or be just /path.
	// We only want to use path part for rewriting and therefore trim prefix if it exists
	rawURI := req.RequestURI
	if rawURI != "" && rawURI[0] != '/' {
		prefix := ""
		if req.URL.Scheme != "" {
			prefix = req.URL.Scheme + "://"
		}
		if req.URL.Host != "" {
			prefix += req.URL.Host // host or host:port
		}
		if prefix != "" {
			rawURI = strings.TrimPrefix(rawURI, prefix)
		}
	}

	for k, v := range rewriteRegex {
		if replacer := captureTokens(k, rawURI); replacer != nil {
			if routed {
				return setRewriteTargetLiterally(req, replacer.Replace(v))
			}
			url, err := req.URL.Parse(replacer.Replace(v))
			if err != nil {
				return err
			}
			req.URL = url

			return nil // rewrite only once
		}
	}
	return nil
}

// setRewriteTargetLiterally sets the path and query of req.URL to target without resolving it.
func setRewriteTargetLiterally(req *http.Request, target string) error {
	target, _, _ = strings.Cut(target, "#") // a fragment is never sent, as with req.URL.Parse
	rawPath, rawQuery, _ := strings.Cut(target, "?")
	decoded, err := url.PathUnescape(rawPath)
	if err != nil || !strings.HasPrefix(decoded, "/") || strings.HasPrefix(decoded, "//") || hasDotSegment(decoded) {
		return echo.ErrBadRequest
	}
	u := *req.URL
	u.Path = decoded
	u.RawPath = ""
	if u.EscapedPath() != rawPath {
		u.RawPath = rawPath // EscapedPath uses RawPath only when it is a valid encoding of Path
	}
	u.RawQuery = rawQuery
	u.Fragment = ""
	u.RawFragment = ""
	req.URL = &u
	return nil
}

// hasDotSegment reports whether decoded URL path p has a "." or ".." segment.
func hasDotSegment(p string) bool {
	for segment := range strings.SplitSeq(p, "/") {
		if segment == "." || segment == ".." {
			return true
		}
	}
	return false
}

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(echo.Context) bool {
	return false
}

func toMiddlewareOrPanic(config interface {
	ToMiddleware() (echo.MiddlewareFunc, error)
}) echo.MiddlewareFunc {
	mw, err := config.ToMiddleware()
	if err != nil {
		panic(err)
	}
	return mw
}
