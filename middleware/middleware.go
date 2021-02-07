package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type (
	// Skipper defines a function to skip middleware. Returning true skips processing
	// the middleware.
	Skipper func(echo.Context) bool

	// BeforeFunc defines a function which is executed just before the middleware.
	BeforeFunc func(echo.Context)
)

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
		k = strings.Replace(k, `\*`, "(.*?)", -1)
		if strings.HasPrefix(k, `\^`) {
			k = strings.Replace(k, `\^`, "^", -1)
		}
		k = k + "$"
		rulesRegex[regexp.MustCompile(k)] = v
	}
	return rulesRegex
}

func rewritePath(rewriteRegex map[*regexp.Regexp]string, req *http.Request) {
	for k, v := range rewriteRegex {
		replacerRawPath := captureTokens(k, req.URL.EscapedPath())
		if replacerRawPath != nil {
			replacerPath := captureTokens(k, req.URL.Path)
			req.URL.RawPath, req.URL.Path = replacerRawPath.Replace(v), replacerPath.Replace(v)
		}
	}
}

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(echo.Context) bool {
	return false
}
