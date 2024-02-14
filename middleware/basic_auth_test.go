package middleware

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuthWithConfig(t *testing.T) {
	validatorFunc := func(u, p string, c echo.Context) (bool, error) {
		if u == "joe" && p == "secret" {
			return true, nil
		}
		if u == "error" {
			return false, errors.New(p)
		}
		return false, nil
	}
	defaultConfig := BasicAuthConfig{Validator: validatorFunc}

	// we can not add OK value here because ranging over map returns random order. We just try to trigger break
	tooManyAuths := make([]string, 0)
	for i := 0; i < extractorLimit+2; i++ {
		tooManyAuths = append(tooManyAuths, basic+" "+base64.StdEncoding.EncodeToString([]byte("nope:nope")))
	}

	var testCases = []struct {
		name         string
		givenConfig  BasicAuthConfig
		whenAuth     []string
		expectHeader string
		expectErr    string
	}{
		{
			name:        "ok",
			givenConfig: defaultConfig,
			whenAuth:    []string{basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))},
		},
		{
			name:        "ok, from multiple auth headers one is ok",
			givenConfig: defaultConfig,
			whenAuth: []string{
				"Bearer " + base64.StdEncoding.EncodeToString([]byte("token")), // different type
				basic + " NOT_BASE64", // invalid basic auth
				basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret")), // OK
			},
		},
		{
			name:         "nok, invalid Authorization header",
			givenConfig:  defaultConfig,
			whenAuth:     []string{strings.ToUpper(basic) + " " + base64.StdEncoding.EncodeToString([]byte("invalid"))},
			expectHeader: basic + ` realm=Restricted`,
			expectErr:    "code=401, message=Unauthorized",
		},
		{
			name:        "nok, not base64 Authorization header",
			givenConfig: defaultConfig,
			whenAuth:    []string{strings.ToUpper(basic) + " NOT_BASE64"},
			expectErr:   "code=400, message=Bad Request, internal=illegal base64 data at input byte 3",
		},
		{
			name:         "nok, missing Authorization header",
			givenConfig:  defaultConfig,
			expectHeader: basic + ` realm=Restricted`,
			expectErr:    "code=401, message=Unauthorized",
		},
		{
			name:         "nok, too many invalid Authorization header",
			givenConfig:  defaultConfig,
			whenAuth:     tooManyAuths,
			expectHeader: basic + ` realm=Restricted`,
			expectErr:    "code=401, message=Unauthorized",
		},
		{
			name:        "ok, realm",
			givenConfig: BasicAuthConfig{Validator: validatorFunc, Realm: "someRealm"},
			whenAuth:    []string{basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))},
		},
		{
			name:        "ok, realm, case-insensitive header scheme",
			givenConfig: BasicAuthConfig{Validator: validatorFunc, Realm: "someRealm"},
			whenAuth:    []string{strings.ToUpper(basic) + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))},
		},
		{
			name:         "nok, realm, invalid Authorization header",
			givenConfig:  BasicAuthConfig{Validator: validatorFunc, Realm: "someRealm"},
			whenAuth:     []string{strings.ToUpper(basic) + " " + base64.StdEncoding.EncodeToString([]byte("invalid"))},
			expectHeader: basic + ` realm="someRealm"`,
			expectErr:    "code=401, message=Unauthorized",
		},
		{
			name:        "nok, validator func returns an error",
			givenConfig: defaultConfig,
			whenAuth:    []string{strings.ToUpper(basic) + " " + base64.StdEncoding.EncodeToString([]byte("error:my_error"))},
			expectErr:   "my_error",
		},
		{
			name: "ok, skipped",
			givenConfig: BasicAuthConfig{Validator: validatorFunc, Skipper: func(c echo.Context) bool {
				return true
			}},
			whenAuth: []string{strings.ToUpper(basic) + " " + base64.StdEncoding.EncodeToString([]byte("invalid"))},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			mw := BasicAuthWithConfig(tc.givenConfig)

			h := mw(func(c echo.Context) error {
				return c.String(http.StatusTeapot, "test")
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			res := httptest.NewRecorder()

			if len(tc.whenAuth) != 0 {
				for _, a := range tc.whenAuth {
					req.Header.Add(echo.HeaderAuthorization, a)
				}
			}
			err := h(e.NewContext(req, res))

			if tc.expectErr != "" {
				assert.Equal(t, http.StatusOK, res.Code)
				assert.EqualError(t, err, tc.expectErr)
			} else {
				assert.Equal(t, http.StatusTeapot, res.Code)
				assert.NoError(t, err)
			}
			if tc.expectHeader != "" {
				assert.Equal(t, tc.expectHeader, res.Header().Get(echo.HeaderWWWAuthenticate))
			}
		})
	}
}

func TestBasicAuth(t *testing.T) {
	e := echo.New()
	f := func(u, p string, c echo.Context) (bool, error) {
		if u == "joe" && p == "secret" {
			return true, nil
		}
		return false, nil
	}
	h := BasicAuth(f)(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)

	auth := basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.HeaderAuthorization, auth)
	assert.NoError(t, h(c))
}
