package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"github.com/stretchr/testify/assert"
)

func TestCSRF_tokenExtractors(t *testing.T) {
	var testCases = []struct {
		name              string
		whenTokenLookup   string
		whenCookieName    string
		givenCSRFCookie   string
		givenMethod       string
		givenQueryTokens  map[string][]string
		givenFormTokens   map[string][]string
		givenHeaderTokens map[string][]string
		expectError       string
	}{
		{
			name:            "ok, multiple token lookups sources, succeeds on last one",
			whenTokenLookup: "header:X-CSRF-Token,form:csrf",
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPost,
			givenHeaderTokens: map[string][]string{
				echo.HeaderXCSRFToken: {"invalid_token"},
			},
			givenFormTokens: map[string][]string{
				"csrf": {"token"},
			},
		},
		{
			name:            "ok, token from POST form",
			whenTokenLookup: "form:csrf",
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPost,
			givenFormTokens: map[string][]string{
				"csrf": {"token"},
			},
		},
		{
			name:            "ok, token from POST form, second token passes",
			whenTokenLookup: "form:csrf",
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPost,
			givenFormTokens: map[string][]string{
				"csrf": {"invalid", "token"},
			},
		},
		{
			name:            "nok, invalid token from POST form",
			whenTokenLookup: "form:csrf",
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPost,
			givenFormTokens: map[string][]string{
				"csrf": {"invalid_token"},
			},
			expectError: "code=403, message=invalid csrf token",
		},
		{
			name:            "nok, missing token from POST form",
			whenTokenLookup: "form:csrf",
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPost,
			givenFormTokens: map[string][]string{},
			expectError:     "code=400, message=missing csrf token in the form parameter",
		},
		{
			name:            "ok, token from POST header",
			whenTokenLookup: "", // will use defaults
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPost,
			givenHeaderTokens: map[string][]string{
				echo.HeaderXCSRFToken: {"token"},
			},
		},
		{
			name:            "ok, token from POST header, second token passes",
			whenTokenLookup: "header:" + echo.HeaderXCSRFToken,
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPost,
			givenHeaderTokens: map[string][]string{
				echo.HeaderXCSRFToken: {"invalid", "token"},
			},
		},
		{
			name:            "nok, invalid token from POST header",
			whenTokenLookup: "header:" + echo.HeaderXCSRFToken,
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPost,
			givenHeaderTokens: map[string][]string{
				echo.HeaderXCSRFToken: {"invalid_token"},
			},
			expectError: "code=403, message=invalid csrf token",
		},
		{
			name:              "nok, missing token from POST header",
			whenTokenLookup:   "header:" + echo.HeaderXCSRFToken,
			givenCSRFCookie:   "token",
			givenMethod:       http.MethodPost,
			givenHeaderTokens: map[string][]string{},
			expectError:       "code=400, message=missing csrf token in request header",
		},
		{
			name:            "ok, token from PUT query param",
			whenTokenLookup: "query:csrf-param",
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPut,
			givenQueryTokens: map[string][]string{
				"csrf-param": {"token"},
			},
		},
		{
			name:            "ok, token from PUT query form, second token passes",
			whenTokenLookup: "query:csrf",
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPut,
			givenQueryTokens: map[string][]string{
				"csrf": {"invalid", "token"},
			},
		},
		{
			name:            "nok, invalid token from PUT query form",
			whenTokenLookup: "query:csrf",
			givenCSRFCookie: "token",
			givenMethod:     http.MethodPut,
			givenQueryTokens: map[string][]string{
				"csrf": {"invalid_token"},
			},
			expectError: "code=403, message=invalid csrf token",
		},
		{
			name:             "nok, missing token from PUT query form",
			whenTokenLookup:  "query:csrf",
			givenCSRFCookie:  "token",
			givenMethod:      http.MethodPut,
			givenQueryTokens: map[string][]string{},
			expectError:      "code=400, message=missing csrf token in the query string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			q := make(url.Values)
			for queryParam, values := range tc.givenQueryTokens {
				for _, v := range values {
					q.Add(queryParam, v)
				}
			}

			f := make(url.Values)
			for formKey, values := range tc.givenFormTokens {
				for _, v := range values {
					f.Add(formKey, v)
				}
			}

			var req *http.Request
			switch tc.givenMethod {
			case http.MethodGet:
				req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
			case http.MethodPost, http.MethodPut:
				req = httptest.NewRequest(http.MethodPost, "/?"+q.Encode(), strings.NewReader(f.Encode()))
				req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
			}

			for header, values := range tc.givenHeaderTokens {
				for _, v := range values {
					req.Header.Add(header, v)
				}
			}

			if tc.givenCSRFCookie != "" {
				req.Header.Set(echo.HeaderCookie, "_csrf="+tc.givenCSRFCookie)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			csrf := CSRFWithConfig(CSRFConfig{
				TokenLookup: tc.whenTokenLookup,
				CookieName:  tc.whenCookieName,
			})

			h := csrf(func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})

			err := h(c)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCSRF(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	csrf := CSRF()
	h := csrf(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Generate CSRF token
	h(c)
	assert.Contains(t, rec.Header().Get(echo.HeaderSetCookie), "_csrf")

	// Without CSRF cookie
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	assert.Error(t, h(c))

	// Empty/invalid CSRF token
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	req.Header.Set(echo.HeaderXCSRFToken, "")
	assert.Error(t, h(c))

	// Valid CSRF token
	token := random.String(32)
	req.Header.Set(echo.HeaderCookie, "_csrf="+token)
	req.Header.Set(echo.HeaderXCSRFToken, token)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestCSRFSetSameSiteMode(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	csrf := CSRFWithConfig(CSRFConfig{
		CookieSameSite: http.SameSiteStrictMode,
	})

	h := csrf(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	r := h(c)
	assert.NoError(t, r)
	assert.Regexp(t, "SameSite=Strict", rec.Header()["Set-Cookie"])
}

func TestCSRFWithoutSameSiteMode(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	csrf := CSRFWithConfig(CSRFConfig{})

	h := csrf(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	r := h(c)
	assert.NoError(t, r)
	assert.NotRegexp(t, "SameSite=", rec.Header()["Set-Cookie"])
}

func TestCSRFWithSameSiteDefaultMode(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	csrf := CSRFWithConfig(CSRFConfig{
		CookieSameSite: http.SameSiteDefaultMode,
	})

	h := csrf(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	r := h(c)
	assert.NoError(t, r)
	assert.NotRegexp(t, "SameSite=", rec.Header()["Set-Cookie"])
}

func TestCSRFWithSameSiteModeNone(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	csrf := CSRFWithConfig(CSRFConfig{
		CookieSameSite: http.SameSiteNoneMode,
	})

	h := csrf(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	r := h(c)
	assert.NoError(t, r)
	assert.Regexp(t, "SameSite=None", rec.Header()["Set-Cookie"])
	assert.Regexp(t, "Secure", rec.Header()["Set-Cookie"])
}

func TestCSRFConfig_skipper(t *testing.T) {
	var testCases = []struct {
		name          string
		whenSkip      bool
		expectCookies int
	}{
		{
			name:          "do skip",
			whenSkip:      true,
			expectCookies: 0,
		},
		{
			name:          "do not skip",
			whenSkip:      false,
			expectCookies: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			csrf := CSRFWithConfig(CSRFConfig{
				Skipper: func(c echo.Context) bool {
					return tc.whenSkip
				},
			})

			h := csrf(func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})

			r := h(c)
			assert.NoError(t, r)
			cookie := rec.Header()["Set-Cookie"]
			assert.Len(t, cookie, tc.expectCookies)
		})
	}
}

func TestCSRFErrorHandling(t *testing.T) {
	cfg := CSRFConfig{
		ErrorHandler: func(err error, c echo.Context) error {
			return echo.NewHTTPError(http.StatusTeapot, "error_handler_executed")
		},
	}

	e := echo.New()
	e.POST("/", func(c echo.Context) error {
		return c.String(http.StatusNotImplemented, "should not end up here")
	})

	e.Use(CSRFWithConfig(cfg))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusTeapot, res.Code)
	assert.Equal(t, "{\"message\":\"error_handler_executed\"}\n", res.Body.String())
}
