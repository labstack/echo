// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"fmt"
	"net/textproto"
	"strings"

	"github.com/labstack/echo/v5"
)

const (
	// extractorLimit is arbitrary number to limit values extractor can return. this limits possible resource exhaustion
	// attack vector
	extractorLimit = 20
)

// ExtractorSource is type to indicate source for extracted value
type ExtractorSource string

const (
	// ExtractorSourceHeader means value was extracted from request header
	ExtractorSourceHeader ExtractorSource = "header"
	// ExtractorSourceQuery means value was extracted from request query parameters
	ExtractorSourceQuery ExtractorSource = "query"
	// ExtractorSourcePathParam means value was extracted from route path parameters
	ExtractorSourcePathParam ExtractorSource = "param"
	// ExtractorSourceCookie means value was extracted from request cookies
	ExtractorSourceCookie ExtractorSource = "cookie"
	// ExtractorSourceForm means value was extracted from request form values
	ExtractorSourceForm ExtractorSource = "form"
)

// ValueExtractorError is error type when middleware extractor is unable to extract value from lookups
type ValueExtractorError struct {
	message string
}

// Error returns errors text
func (e *ValueExtractorError) Error() string {
	return e.message
}

var errHeaderExtractorValueMissing = &ValueExtractorError{message: "missing value in request header"}
var errHeaderExtractorValueInvalid = &ValueExtractorError{message: "invalid value in request header"}
var errQueryExtractorValueMissing = &ValueExtractorError{message: "missing value in the query string"}
var errParamExtractorValueMissing = &ValueExtractorError{message: "missing value in path params"}
var errCookieExtractorValueMissing = &ValueExtractorError{message: "missing value in cookies"}
var errFormExtractorValueMissing = &ValueExtractorError{message: "missing value in the form"}

// ValuesExtractor defines a function for extracting values (keys/tokens) from the given context.
type ValuesExtractor func(c *echo.Context) ([]string, ExtractorSource, error)

// CreateExtractors creates ValuesExtractors from given lookups.
// lookups is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
// to extract key from the request.
// Possible values:
//   - "header:<name>" or "header:<name>:<cut-prefix>"
//     `<cut-prefix>` is argument value to cut/trim prefix of the extracted value. This is useful if header
//     value has static prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we
//     want to cut is `<auth-scheme> ` note the space at the end.
//     In case of basic authentication `Authorization: Basic <credentials>` prefix we want to remove is `Basic `.
//   - "query:<name>"
//   - "param:<name>"
//   - "form:<name>"
//   - "cookie:<name>"
//
// Multiple sources example:
// - "header:Authorization,header:X-Api-Key"
//
// limit sets the maximum amount how many lookups can be returned.
func CreateExtractors(lookups string, limit uint) ([]ValuesExtractor, error) {
	return createExtractors(lookups, limit)
}

func createExtractors(lookups string, limit uint) ([]ValuesExtractor, error) {
	if lookups == "" {
		return nil, nil
	}
	if limit == 0 {
		limit = 1
	} else if limit > extractorLimit {
		limit = extractorLimit
	}

	sources := strings.Split(lookups, ",")
	var extractors = make([]ValuesExtractor, 0)
	for _, source := range sources {
		parts := strings.Split(source, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("extractor source for lookup could not be split into needed parts: %v", source)
		}

		switch parts[0] {
		case "query":
			extractors = append(extractors, valuesFromQuery(parts[1], limit))
		case "param":
			extractors = append(extractors, valuesFromParam(parts[1], limit))
		case "cookie":
			extractors = append(extractors, valuesFromCookie(parts[1], limit))
		case "form":
			extractors = append(extractors, valuesFromForm(parts[1], limit))
		case "header":
			prefix := ""
			if len(parts) > 2 {
				prefix = parts[2]
			}
			extractors = append(extractors, valuesFromHeader(parts[1], prefix, limit))
		}
	}
	return extractors, nil
}

// valuesFromHeader returns a functions that extracts values from the request header.
// valuePrefix is parameter to remove first part (prefix) of the extracted value. This is useful if header value has static
// prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we want to remove is `<auth-scheme> `
// note the space at the end. In case of basic authentication `Authorization: Basic <credentials>` prefix we want to remove
// is `Basic `. In case of JWT tokens `Authorization: Bearer <token>` prefix is `Bearer `.
// If prefix is left empty the whole value is returned.
func valuesFromHeader(header string, valuePrefix string, limit uint) ValuesExtractor {
	prefixLen := len(valuePrefix)
	// standard library parses http.Request header keys in canonical form but we may provide something else so fix this
	header = textproto.CanonicalMIMEHeaderKey(header)
	if limit == 0 {
		limit = 1
	}
	return func(c *echo.Context) ([]string, ExtractorSource, error) {
		values := c.Request().Header.Values(header)
		if len(values) == 0 {
			return nil, ExtractorSourceHeader, errHeaderExtractorValueMissing
		}

		i := uint(0)
		result := make([]string, 0)
		for _, value := range values {
			if prefixLen == 0 {
				result = append(result, value)
				i++
				if i >= limit {
					break
				}
			} else if len(value) > prefixLen && strings.EqualFold(value[:prefixLen], valuePrefix) {
				result = append(result, value[prefixLen:])
				i++
				if i >= limit {
					break
				}
			}
		}

		if len(result) == 0 {
			if prefixLen > 0 {
				return nil, ExtractorSourceHeader, errHeaderExtractorValueInvalid
			}
			return nil, ExtractorSourceHeader, errHeaderExtractorValueMissing
		}
		return result, ExtractorSourceHeader, nil
	}
}

// valuesFromQuery returns a function that extracts values from the query string.
func valuesFromQuery(param string, limit uint) ValuesExtractor {
	if limit == 0 {
		limit = 1
	}
	return func(c *echo.Context) ([]string, ExtractorSource, error) {
		result := c.QueryParams()[param]
		if len(result) == 0 {
			return nil, ExtractorSourceQuery, errQueryExtractorValueMissing
		} else if len(result) > int(limit)-1 {
			result = result[:limit]
		}
		return result, ExtractorSourceQuery, nil
	}
}

// valuesFromParam returns a function that extracts values from the url param string.
func valuesFromParam(param string, limit uint) ValuesExtractor {
	if limit == 0 {
		limit = 1
	}
	return func(c *echo.Context) ([]string, ExtractorSource, error) {
		result := make([]string, 0)
		i := uint(0)
		for _, p := range c.PathValues() {
			if param != p.Name {
				continue
			}
			result = append(result, p.Value)
			i++
			if i >= limit {
				break
			}
		}
		if len(result) == 0 {
			return nil, ExtractorSourcePathParam, errParamExtractorValueMissing
		}
		return result, ExtractorSourcePathParam, nil
	}
}

// valuesFromCookie returns a function that extracts values from the named cookie.
func valuesFromCookie(name string, limit uint) ValuesExtractor {
	if limit == 0 {
		limit = 1
	}
	return func(c *echo.Context) ([]string, ExtractorSource, error) {
		cookies := c.Cookies()
		if len(cookies) == 0 {
			return nil, ExtractorSourceCookie, errCookieExtractorValueMissing
		}

		i := uint(0)
		result := make([]string, 0)
		for _, cookie := range cookies {
			if name != cookie.Name {
				continue
			}
			result = append(result, cookie.Value)
			i++
			if i >= limit {
				break
			}
		}
		if len(result) == 0 {
			return nil, ExtractorSourceCookie, errCookieExtractorValueMissing
		}
		return result, ExtractorSourceCookie, nil
	}
}

// valuesFromForm returns a function that extracts values from the form field.
func valuesFromForm(name string, limit uint) ValuesExtractor {
	if limit == 0 {
		limit = 1
	}
	return func(c *echo.Context) ([]string, ExtractorSource, error) {
		if c.Request().Form == nil {
			_, _ = c.MultipartForm() // we want to trigger c.request.ParseMultipartForm(c.formParseMaxMemory)
		}
		values := c.Request().Form[name]
		if len(values) == 0 {
			return nil, ExtractorSourceForm, errFormExtractorValueMissing
		}
		if len(values) > int(limit)-1 {
			values = values[:limit]
		}
		result := append([]string{}, values...)
		return result, ExtractorSourceForm, nil
	}
}
