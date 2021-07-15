package middleware

import (
	"fmt"
	"github.com/labstack/echo/v5"
	"net/textproto"
	"strings"
)

const (
	// extractorLimit is arbitrary number to limit values extractor can return. this limits possible resource exhaustion
	// attack vector
	extractorLimit = 20
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
type ValuesExtractor func(c echo.Context) ([]string, error)

func createExtractors(lookups string) ([]ValuesExtractor, error) {
	if lookups == "" {
		return nil, nil
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
			extractors = append(extractors, valuesFromQuery(parts[1]))
		case "param":
			extractors = append(extractors, valuesFromParam(parts[1]))
		case "cookie":
			extractors = append(extractors, valuesFromCookie(parts[1]))
		case "form":
			extractors = append(extractors, valuesFromForm(parts[1]))
		case "header":
			prefix := ""
			if len(parts) > 2 {
				prefix = parts[2]
			}
			extractors = append(extractors, valuesFromHeader(parts[1], prefix))
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
func valuesFromHeader(header string, valuePrefix string) ValuesExtractor {
	prefixLen := len(valuePrefix)
	// standard library parses http.Request header keys in canonical form but we may provide something else so fix this
	header = textproto.CanonicalMIMEHeaderKey(header)
	return func(c echo.Context) ([]string, error) {
		values := c.Request().Header.Values(header)
		if len(values) == 0 {
			return nil, errHeaderExtractorValueMissing
		}

		result := make([]string, 0)
		for i, value := range values {
			if prefixLen == 0 {
				result = append(result, value)
				if i >= extractorLimit-1 {
					break
				}
				continue
			}
			if len(value) > prefixLen && strings.EqualFold(value[:prefixLen], valuePrefix) {
				result = append(result, value[prefixLen:])
				if i >= extractorLimit-1 {
					break
				}
			}
		}

		if len(result) == 0 {
			if prefixLen > 0 {
				return nil, errHeaderExtractorValueInvalid
			}
			return nil, errHeaderExtractorValueMissing
		}
		return result, nil
	}
}

// valuesFromQuery returns a function that extracts values from the query string.
func valuesFromQuery(param string) ValuesExtractor {
	return func(c echo.Context) ([]string, error) {
		result := c.QueryParams()[param]
		if len(result) == 0 {
			return nil, errQueryExtractorValueMissing
		} else if len(result) > extractorLimit-1 {
			result = result[:extractorLimit]
		}
		return result, nil
	}
}

// valuesFromParam returns a function that extracts values from the url param string.
func valuesFromParam(param string) ValuesExtractor {
	return func(c echo.Context) ([]string, error) {
		result := make([]string, 0)
		for i, p := range c.PathParams() {
			if param == p.Name {
				result = append(result, p.Value)
				if i >= extractorLimit-1 {
					break
				}
			}
		}
		if len(result) == 0 {
			return nil, errParamExtractorValueMissing
		}
		return result, nil
	}
}

// valuesFromCookie returns a function that extracts values from the named cookie.
func valuesFromCookie(name string) ValuesExtractor {
	return func(c echo.Context) ([]string, error) {
		cookies := c.Cookies()
		if len(cookies) == 0 {
			return nil, errCookieExtractorValueMissing
		}

		result := make([]string, 0)
		for i, cookie := range cookies {
			if name == cookie.Name {
				result = append(result, cookie.Value)
				if i >= extractorLimit-1 {
					break
				}
			}
		}
		if len(result) == 0 {
			return nil, errCookieExtractorValueMissing
		}
		return result, nil
	}
}

// valuesFromForm returns a function that extracts values from the form field.
func valuesFromForm(name string) ValuesExtractor {
	return func(c echo.Context) ([]string, error) {
		if parseErr := c.Request().ParseForm(); parseErr != nil {
			return nil, fmt.Errorf("valuesFromForm parse form failed: %w", parseErr)
		}
		values := c.Request().Form[name]
		if len(values) == 0 {
			return nil, errFormExtractorValueMissing
		}
		if len(values) > extractorLimit-1 {
			values = values[:extractorLimit]
		}
		result := append([]string{}, values...)
		return result, nil
	}
}
