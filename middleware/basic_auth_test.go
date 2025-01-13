// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

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

func TestBasicAuth(t *testing.T) {

	validator := func(u, p string, c echo.Context) (bool, error) {
		if u == "joe" && p == "secret" {
			return true, nil
		}
		return false, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	userPassB64 := base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(echo.HeaderAuthorization, basic+" "+userPassB64)

	e := echo.New()
	c := e.NewContext(req, res)

	h := BasicAuth(validator)(func(c echo.Context) error {
		return c.String(http.StatusIMUsed, "test")
	})
	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusIMUsed, res.Code)
}

func TestBasicAuthPanic(t *testing.T) {
	assert.PanicsWithError(t, "echo basic-auth middleware requires a validator function", func() {
		BasicAuth(nil)
	})
}

func TestBasicAuthWithConfig(t *testing.T) {
	e := echo.New()

	exampleSecret := base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	mockValidator := func(u, p string, c echo.Context) (bool, error) {
		if u == "error" {
			return false, errors.New("validator_error")
		}
		if u == "joe" && p == "secret" {
			return true, nil
		}
		return false, nil
	}

	tests := []struct {
		name           string
		authHeader     []string
		config         *BasicAuthConfig
		expectedCode   int
		expectedAuth   string
		expectedErr    string
		expectedErrMsg string
	}{
		{
			name:         "Valid credentials",
			authHeader:   []string{basic + " " + exampleSecret},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Case-insensitive header scheme",
			authHeader:   []string{strings.ToUpper(basic) + " " + exampleSecret},
			expectedCode: http.StatusOK,
		},
		{
			name:           "Invalid credentials",
			authHeader:     []string{basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:invalid-password"))},
			expectedCode:   http.StatusUnauthorized,
			expectedAuth:   basic + ` realm="someRealm"`,
			expectedErr:    "code=401, message=Unauthorized",
			expectedErrMsg: "Unauthorized",
		},
		{
			name: "validator errors out at 2 tries",
			authHeader: []string{
				basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:invalid-password")),
				basic + " " + base64.StdEncoding.EncodeToString([]byte("error:secret")),
			},
			config: &BasicAuthConfig{
				HeaderValidationLimit: 2,
				Validator:             mockValidator,
			},
			expectedCode:   http.StatusUnauthorized,
			expectedAuth:   "",
			expectedErr:    "validator_error",
			expectedErrMsg: "Unauthorized",
		},
		{
			name:         "Invalid credentials, default realm",
			authHeader:   []string{basic + " " + exampleSecret},
			expectedCode: http.StatusOK,
			expectedAuth: basic + ` realm="Restricted"`,
		},
		{
			name:           "Invalid base64 string",
			authHeader:     []string{basic + " invalidString"},
			expectedCode:   http.StatusBadRequest,
			expectedErr:    "code=400, message=Bad Request, internal=illegal base64 data at input byte 12",
			expectedErrMsg: "Bad Request",
		},
		{
			name:           "Missing Authorization header",
			expectedCode:   http.StatusUnauthorized,
			expectedErr:    "code=401, message=Unauthorized",
			expectedErrMsg: "Unauthorized",
		},
		{
			name:           "Invalid Authorization header",
			authHeader:     []string{base64.StdEncoding.EncodeToString([]byte("invalid"))},
			expectedCode:   http.StatusUnauthorized,
			expectedErr:    "code=401, message=Unauthorized",
			expectedErrMsg: "Unauthorized",
		},
		{
			name:         "Skipped Request",
			authHeader:   []string{basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:skip"))},
			expectedCode: http.StatusOK,
			config: &BasicAuthConfig{
				Validator: mockValidator,
				Realm:     "someRealm",
				Skipper: func(c echo.Context) bool {
					return true
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			for _, h := range tt.authHeader {
				req.Header.Add(echo.HeaderAuthorization, h)
			}

			config := BasicAuthConfig{
				Validator: mockValidator,
				Realm:     "someRealm",
			}
			if tt.config != nil {
				config = *tt.config
			}
			h := BasicAuthWithConfig(config)(func(c echo.Context) error {
				return c.String(http.StatusOK, "test")
			})

			err := h(c)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
				if tt.expectedAuth != "" {
					assert.Equal(t, tt.expectedAuth, res.Header().Get(echo.HeaderWWWAuthenticate))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, res.Code)
			}
		})
	}
}
