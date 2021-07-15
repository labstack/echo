package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func createTestParseTokenFuncForJWTGo(signingMethod string, signingKey interface{}) func(c echo.Context, auth string) (interface{}, error) {
	// This is minimal implementation for github.com/golang-jwt/jwt as JWT parser library. good enough to get old tests running
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != signingMethod {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return signingKey, nil
	}

	return func(c echo.Context, auth string) (interface{}, error) {
		token, err := jwt.ParseWithClaims(auth, jwt.MapClaims{}, keyFunc)
		if err != nil {
			return nil, err
		}
		if !token.Valid {
			return nil, errors.New("invalid token")
		}
		return token, nil
	}
}

// jwtCustomInfo defines some custom types we're going to use within our tokens.
type jwtCustomInfo struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}

// jwtCustomClaims are custom claims expanding default ones.
type jwtCustomClaims struct {
	*jwt.StandardClaims
	jwtCustomInfo
}

func TestJWT_combinations(t *testing.T) {
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
	validKey := []byte("secret")
	invalidKey := []byte("invalid-key")
	validAuth := "Bearer " + token

	var testCases = []struct {
		name       string
		config     JWTConfig
		reqURL     string // "/" if empty
		hdrAuth    string
		hdrCookie  string // test.Request doesn't provide SetCookie(); use name=val
		formValues map[string]string
		expPanic   bool
		expErrCode int // 0 for Success
	}{
		{
			expPanic: true,
			name:     "No signing key provided",
		},
		{
			expErrCode: http.StatusUnauthorized,
			hdrAuth:    validAuth,
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo("RS256", validKey),
			},
			name: "Unexpected signing method",
		},
		{
			expErrCode: http.StatusUnauthorized,
			hdrAuth:    validAuth,
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, invalidKey),
			},
			name: "Invalid key",
		},
		{
			hdrAuth: validAuth,
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
			},
			name: "Valid JWT",
		},
		{
			hdrAuth: "Token" + " " + token,
			config: JWTConfig{
				TokenLookup:    "header:" + echo.HeaderAuthorization + ":Token ",
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
			},
			name: "Valid JWT with custom AuthScheme",
		},
		{
			hdrAuth: validAuth,
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, []byte("secret")),
			},
			name: "Valid JWT with custom claims",
		},
		{
			hdrAuth:    "invalid-auth",
			expErrCode: http.StatusUnauthorized,
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
			},
			name: "Invalid Authorization header",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
			},
			expErrCode: http.StatusUnauthorized,
			name:       "Empty header auth field",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "query:jwt",
			},
			reqURL: "/?a=b&jwt=" + token,
			name:   "Valid query method",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "query:jwt",
			},
			reqURL:     "/?a=b&jwtxyz=" + token,
			expErrCode: http.StatusUnauthorized,
			name:       "Invalid query param name",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "query:jwt",
			},
			reqURL:     "/?a=b&jwt=invalid-token",
			expErrCode: http.StatusUnauthorized,
			name:       "Invalid query param value",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "query:jwt",
			},
			reqURL:     "/?a=b",
			expErrCode: http.StatusUnauthorized,
			name:       "Empty query",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "param:jwt",
			},
			reqURL: "/" + token,
			name:   "Valid param method",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "cookie:jwt",
			},
			hdrCookie: "jwt=" + token,
			name:      "Valid cookie method",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "query:jwt,cookie:jwt",
			},
			hdrCookie: "jwt=" + token,
			name:      "Multiple jwt lookuop",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "cookie:jwt",
			},
			expErrCode: http.StatusUnauthorized,
			hdrCookie:  "jwt=invalid",
			name:       "Invalid token with cookie method",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "cookie:jwt",
			},
			expErrCode: http.StatusUnauthorized,
			name:       "Empty cookie",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "form:jwt",
			},
			formValues: map[string]string{"jwt": token},
			name:       "Valid form method",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "form:jwt",
			},
			expErrCode: http.StatusUnauthorized,
			formValues: map[string]string{"jwt": "invalid"},
			name:       "Invalid token with form method",
		},
		{
			config: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, validKey),
				TokenLookup:    "form:jwt",
			},
			expErrCode: http.StatusUnauthorized,
			name:       "Empty form field",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.reqURL == "" {
				tc.reqURL = "/"
			}

			var req *http.Request
			if len(tc.formValues) > 0 {
				form := url.Values{}
				for k, v := range tc.formValues {
					form.Set(k, v)
				}
				req = httptest.NewRequest(http.MethodPost, tc.reqURL, strings.NewReader(form.Encode()))
				req.Header.Set(echo.HeaderContentType, "application/x-www-form-urlencoded")
				req.ParseForm()
			} else {
				req = httptest.NewRequest(http.MethodGet, tc.reqURL, nil)
			}
			res := httptest.NewRecorder()
			req.Header.Set(echo.HeaderAuthorization, tc.hdrAuth)
			req.Header.Set(echo.HeaderCookie, tc.hdrCookie)
			c := e.NewContext(req, res)

			if tc.reqURL == "/"+token {
				cc := c.(echo.EditableContext)
				cc.SetPathParams(echo.PathParams{
					{Name: "jwt", Value: token},
				})
			}

			if tc.expPanic {
				assert.Panics(t, func() {
					JWTWithConfig(tc.config)
				}, tc.name)
				return
			}

			if tc.expErrCode != 0 {
				h := JWTWithConfig(tc.config)(handler)
				he := h(c).(*echo.HTTPError)
				assert.Equal(t, tc.expErrCode, he.Code)
				return
			}

			h := JWTWithConfig(tc.config)(handler)
			if assert.NoError(t, h(c), tc.name) {
				user := c.Get("user").(*jwt.Token)
				switch claims := user.Claims.(type) {
				case jwt.MapClaims:
					assert.Equal(t, claims["name"], "John Doe")
				case *jwtCustomClaims:
					assert.Equal(t, claims.Name, "John Doe")
					assert.Equal(t, claims.Admin, true)
				default:
					panic("unexpected type of claims")
				}
			}
		})
	}
}

func TestJWTConfig_skipper(t *testing.T) {
	e := echo.New()

	e.Use(JWTWithConfig(JWTConfig{
		Skipper: func(context echo.Context) bool {
			return true // skip everything
		},
		ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, []byte("secret")),
	}))

	isCalled := false
	e.GET("/", func(c echo.Context) error {
		isCalled = true
		return c.String(http.StatusTeapot, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusTeapot, res.Code)
	assert.True(t, isCalled)
}

func TestJWTConfig_BeforeFunc(t *testing.T) {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusTeapot, "test")
	})

	isCalled := false
	e.Use(JWTWithConfig(JWTConfig{
		BeforeFunc: func(context echo.Context) {
			isCalled = true
		},
		ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, []byte("secret")),
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusTeapot, res.Code)
	assert.True(t, isCalled)
}

func TestJWTConfig_extractorErrorHandling(t *testing.T) {
	var testCases = []struct {
		name             string
		given            JWTConfig
		expectStatusCode int
	}{
		{
			name: "ok, ErrorHandler is executed",
			given: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, []byte("secret")),
				ErrorHandler: func(c echo.Context, err error) error {
					return echo.NewHTTPError(http.StatusTeapot, "custom_error")
				},
			},
			expectStatusCode: http.StatusTeapot,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.GET("/", func(c echo.Context) error {
				return c.String(http.StatusNotImplemented, "should not end up here")
			})

			e.Use(JWTWithConfig(tc.given))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			res := httptest.NewRecorder()
			e.ServeHTTP(res, req)

			assert.Equal(t, tc.expectStatusCode, res.Code)
		})
	}
}

func TestJWTConfig_parseTokenErrorHandling(t *testing.T) {
	var testCases = []struct {
		name      string
		given     JWTConfig
		expectErr string
	}{
		{
			name: "ok, ErrorHandler is executed",
			given: JWTConfig{
				ParseTokenFunc: createTestParseTokenFuncForJWTGo(AlgorithmHS256, []byte("secret")),
				ErrorHandler: func(c echo.Context, err error) error {
					return echo.NewHTTPError(http.StatusTeapot, "ErrorHandler: "+err.Error())
				},
			},
			expectErr: "{\"message\":\"ErrorHandler: parsing failed\"}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			//e.Debug = true
			e.GET("/", func(c echo.Context) error {
				return c.String(http.StatusNotImplemented, "should not end up here")
			})

			config := tc.given
			parseTokenCalled := false
			config.ParseTokenFunc = func(c echo.Context, auth string) (interface{}, error) {
				parseTokenCalled = true
				return nil, errors.New("parsing failed")
			}
			e.Use(JWTWithConfig(config))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
			res := httptest.NewRecorder()

			e.ServeHTTP(res, req)

			assert.Equal(t, http.StatusTeapot, res.Code)
			assert.Equal(t, tc.expectErr, res.Body.String())
			assert.True(t, parseTokenCalled)
		})
	}
}

func TestJWTConfig_custom_ParseTokenFunc_Keyfunc(t *testing.T) {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusTeapot, "test")
	})

	// example of minimal custom ParseTokenFunc implementation. Allows you to use different versions of `github.com/golang-jwt/jwt`
	// with current JWT middleware
	signingKey := []byte("secret")

	config := JWTConfig{
		ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
			keyFunc := func(t *jwt.Token) (interface{}, error) {
				if t.Method.Alg() != "HS256" {
					return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
				}
				return signingKey, nil
			}

			// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
			token, err := jwt.Parse(auth, keyFunc)
			if err != nil {
				return nil, err
			}
			if !token.Valid {
				return nil, errors.New("invalid token")
			}
			return token, nil
		},
	}

	e.Use(JWTWithConfig(config))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusTeapot, res.Code)
}

func TestMustJWTWithConfig_SuccessHandler(t *testing.T) {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		success := c.Get("success").(string)
		user := c.Get("user").(string)
		return c.String(http.StatusTeapot, fmt.Sprintf("%v:%v", success, user))
	})

	mw, err := JWTConfig{
		ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
			return auth, nil
		},
		SuccessHandler: func(c echo.Context) {
			c.Set("success", "yes")
		},
	}.ToMiddleware()
	assert.NoError(t, err)
	e.Use(mw)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add(echo.HeaderAuthorization, "Bearer valid_token_base64")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, "yes:valid_token_base64", res.Body.String())
	assert.Equal(t, http.StatusTeapot, res.Code)
}

func TestJWTWithConfig_CallNextOnNilErrorHandlerResult(t *testing.T) {
	var testCases = []struct {
		name                string
		givenCallNext       bool
		givenErrorHandler   JWTErrorHandlerWithContext
		givenTokenLookup    string
		whenAuthHeaders     []string
		whenCookies         []string
		whenParseReturn     string
		whenParseError      error
		expectHandlerCalled bool
		expect              string
		expectCode          int
	}{
		{
			name:          "ok, with valid JWT from auth header",
			givenCallNext: true,
			givenErrorHandler: func(c echo.Context, err error) error {
				return nil
			},
			whenAuthHeaders: []string{"Bearer valid_token_base64"},
			whenParseReturn: "valid_token",
			expectCode:      http.StatusTeapot,
			expect:          "valid_token",
		},
		{
			name:          "ok, missing header, callNext and set public_token from error handler",
			givenCallNext: true,
			givenErrorHandler: func(c echo.Context, err error) error {
				if err != ErrJWTMissing {
					panic("must get ErrJWTMissing")
				}
				c.Set("user", "public_token")
				return nil
			},
			whenAuthHeaders: []string{}, // no JWT header
			expectCode:      http.StatusTeapot,
			expect:          "public_token",
		},
		{
			name:          "ok, invalid token, callNext and set public_token from error handler",
			givenCallNext: true,
			givenErrorHandler: func(c echo.Context, err error) error {
				// this is probably not realistic usecase. on parse error you probably want to return error
				if err.Error() != "parser_error" {
					panic("must get parser_error")
				}
				c.Set("user", "public_token")
				return nil
			},
			whenAuthHeaders: []string{"Bearer invalid_header"},
			whenParseError:  errors.New("parser_error"),
			expectCode:      http.StatusTeapot,
			expect:          "public_token",
		},
		{
			name:          "nok, invalid token, return error from error handler",
			givenCallNext: true,
			givenErrorHandler: func(c echo.Context, err error) error {
				if err.Error() != "parser_error" {
					panic("must get parser_error")
				}
				return err
			},
			whenAuthHeaders: []string{"Bearer invalid_header"},
			whenParseError:  errors.New("parser_error"),
			expectCode:      http.StatusInternalServerError,
			expect:          "{\"message\":\"Internal Server Error\"}\n",
		},
		{
			name:          "nok, callNext but return error from error handler",
			givenCallNext: true,
			givenErrorHandler: func(c echo.Context, err error) error {
				return err
			},
			whenAuthHeaders: []string{}, // no JWT header
			expectCode:      http.StatusUnauthorized,
			expect:          "{\"message\":\"missing or malformed jwt\"}\n",
		},
		{
			name:          "nok, callNext=false",
			givenCallNext: false,
			givenErrorHandler: func(c echo.Context, err error) error {
				return err
			},
			whenAuthHeaders: []string{}, // no JWT header
			expectCode:      http.StatusUnauthorized,
			expect:          "{\"message\":\"missing or malformed jwt\"}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			e.GET("/", func(c echo.Context) error {
				token := c.Get("user").(string)
				return c.String(http.StatusTeapot, token)
			})

			mw, err := JWTConfig{
				TokenLookup: tc.givenTokenLookup,
				ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
					return tc.whenParseReturn, tc.whenParseError
				},
				ErrorHandler: tc.givenErrorHandler,
			}.ToMiddleware()
			assert.NoError(t, err)
			e.Use(mw)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for _, a := range tc.whenAuthHeaders {
				req.Header.Add(echo.HeaderAuthorization, a)
			}
			res := httptest.NewRecorder()
			e.ServeHTTP(res, req)

			assert.Equal(t, tc.expect, res.Body.String())
			assert.Equal(t, tc.expectCode, res.Code)
		})
	}
}
