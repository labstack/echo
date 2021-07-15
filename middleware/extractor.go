package middleware

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/textproto"
	"strings"
)

// ErrExtractionValueMissing denotes an error raised when value could not be extracted from request
var ErrExtractionValueMissing = echo.NewHTTPError(http.StatusBadRequest, "missing or malformed value")

// ExtractorType is enum type for where extractor will take its data
type ExtractorType string

const (
	// HeaderExtractor tells extractor to take values from request header
	HeaderExtractor ExtractorType = "header"
	// QueryExtractor tells extractor to take values from request query parameters
	QueryExtractor ExtractorType = "query"
	// ParamExtractor tells extractor to take values from request route parameters
	ParamExtractor ExtractorType = "param"
	// CookieExtractor tells extractor to take values from request cookie
	CookieExtractor ExtractorType = "cookie"
	// FormExtractor tells extractor to take values from request form fields
	FormExtractor ExtractorType = "form"
)

func createExtractors(lookups string) ([]valuesExtractor, error) {
	sources := strings.Split(lookups, ",")
	var extractors []valuesExtractor
	for _, source := range sources {
		parts := strings.Split(source, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("extractor source for lookup could not be split into needed parts: %v", source)
		}

		switch ExtractorType(parts[0]) {
		case QueryExtractor:
			extractors = append(extractors, valuesFromQuery(parts[1]))
		case ParamExtractor:
			extractors = append(extractors, valuesFromParam(parts[1]))
		case CookieExtractor:
			extractors = append(extractors, valuesFromCookie(parts[1]))
		case FormExtractor:
			extractors = append(extractors, valuesFromForm(parts[1]))
		case HeaderExtractor:
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
func valuesFromHeader(header string, valuePrefix string) valuesExtractor {
	prefixLen := len(valuePrefix)
	return func(c echo.Context) ([]string, ExtractorType, error) {
		values := textproto.MIMEHeader(c.Request().Header).Values(header)
		if len(values) == 0 {
			return nil, HeaderExtractor, ErrExtractionValueMissing
		}

		result := make([]string, 0)
		for _, value := range values {
			if prefixLen == 0 {
				result = append(result, value)
				continue
			}
			if len(value) > prefixLen && strings.EqualFold(value[:prefixLen], valuePrefix) {
				result = append(result, value[prefixLen:])
			}
		}
		if len(result) == 0 {
			return nil, HeaderExtractor, ErrExtractionValueMissing
		}
		return result, HeaderExtractor, nil
	}
}

// valuesFromQuery returns a function that extracts values from the query string.
func valuesFromQuery(param string) valuesExtractor {
	return func(c echo.Context) ([]string, ExtractorType, error) {
		result := c.QueryParams()[param]
		if len(result) == 0 {
			return nil, QueryExtractor, ErrExtractionValueMissing
		}
		return result, QueryExtractor, nil

	}
}

// valuesFromParam returns a function that extracts values from the url param string.
func valuesFromParam(param string) valuesExtractor {
	return func(c echo.Context) ([]string, ExtractorType, error) {
		result := make([]string, 0)
		for _, p := range c.PathParams() {
			if param == p.Name {
				result = append(result, p.Value)
			}
		}
		if len(result) == 0 {
			return nil, ParamExtractor, ErrExtractionValueMissing
		}
		return result, ParamExtractor, nil
	}
}

// valuesFromCookie returns a function that extracts values from the named cookie.
func valuesFromCookie(name string) valuesExtractor {
	return func(c echo.Context) ([]string, ExtractorType, error) {
		cookies := c.Cookies()
		if len(cookies) == 0 {
			return nil, CookieExtractor, ErrExtractionValueMissing
		}

		result := make([]string, 0)
		for _, cookie := range cookies {
			if name == cookie.Name {
				result = append(result, cookie.Value)
			}
		}
		if len(result) == 0 {
			return nil, CookieExtractor, ErrExtractionValueMissing
		}
		return result, CookieExtractor, nil
	}
}

// valuesFromForm returns a function that extracts values from the form field.
func valuesFromForm(name string) valuesExtractor {
	return func(c echo.Context) ([]string, ExtractorType, error) {
		if err := c.Request().ParseForm(); err != nil {
			return nil, FormExtractor, fmt.Errorf("valuesFromForm parse form failed: %w", err)
		}
		values := c.Request().Form[name]
		if len(values) == 0 {
			return nil, FormExtractor, ErrExtractionValueMissing
		}

		result := append([]string{}, values...)
		return result, FormExtractor, nil
	}
}
