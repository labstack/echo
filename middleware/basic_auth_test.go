package middleware

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {
	validatorFunc := func(c echo.Context, u, p string) (bool, error) {
		if u == "joe" && p == "secret" {
			return true, nil
		}
		if u == "error" {
			return false, errors.New(p)
		}
		return false, nil
	}
	defaultConfig := BasicAuthConfig{Validator: validatorFunc}

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
			name:        "ok, multiple",
			givenConfig: defaultConfig,
			whenAuth: []string{
				"Bearer " + base64.StdEncoding.EncodeToString([]byte("token")),
				basic + " NOT_BASE64",
				basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret")),
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
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			config := tc.givenConfig

			mw, err := config.ToMiddleware()
			assert.NoError(t, err)

			h := mw(func(c echo.Context) error {
				return c.String(http.StatusTeapot, "test")
			})

			if len(tc.whenAuth) != 0 {
				for _, a := range tc.whenAuth {
					req.Header.Add(echo.HeaderAuthorization, a)
				}
			}
			err = h(c)

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

func TestBasicAuth_panic(t *testing.T) {
	assert.Panics(t, func() {
		mw := BasicAuth(nil)
		assert.NotNil(t, mw)
	})

	mw := BasicAuth(func(c echo.Context, user string, password string) (bool, error) {
		return true, nil
	})
	assert.NotNil(t, mw)
}

func TestBasicAuthWithConfig_panic(t *testing.T) {
	assert.Panics(t, func() {
		mw := BasicAuthWithConfig(BasicAuthConfig{Validator: nil})
		assert.NotNil(t, mw)
	})

	mw := BasicAuthWithConfig(BasicAuthConfig{Validator: func(c echo.Context, user string, password string) (bool, error) {
		return true, nil
	}})
	assert.NotNil(t, mw)
}
