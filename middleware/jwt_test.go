//go:build go1.15
// +build go1.15

package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

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

func TestJWT(t *testing.T) {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)
		return c.JSON(http.StatusOK, token.Claims)
	})

	e.Use(JWT([]byte("secret")))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
	res := httptest.NewRecorder()

	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"admin":true,"name":"John Doe","sub":"1234567890"}`+"\n", res.Body.String())
}

func TestJWTRace(t *testing.T) {
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	initialToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
	raceToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlJhY2UgQ29uZGl0aW9uIiwiYWRtaW4iOmZhbHNlfQ.Xzkx9mcgGqYMTkuxSCbJ67lsDyk5J2aB7hu65cEE-Ss"
	validKey := []byte("secret")

	h := JWTWithConfig(JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: validKey,
	})(handler)

	makeReq := func(token string) echo.Context {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderAuthorization, DefaultJWTConfig.AuthScheme+" "+token)
		c := e.NewContext(req, res)
		assert.NoError(t, h(c))
		return c
	}

	c := makeReq(initialToken)
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwtCustomClaims)
	assert.Equal(t, claims.Name, "John Doe")

	makeReq(raceToken)
	user = c.Get("user").(*jwt.Token)
	claims = user.Claims.(*jwtCustomClaims)
	// Initial context should still be "John Doe", not "Race Condition"
	assert.Equal(t, claims.Name, "John Doe")
	assert.Equal(t, claims.Admin, true)
}

func TestJWTConfig(t *testing.T) {
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
	validKey := []byte("secret")
	invalidKey := []byte("invalid-key")
	validAuth := DefaultJWTConfig.AuthScheme + " " + token

	testCases := []struct {
		name       string
		expPanic   bool
		expErrCode int // 0 for Success
		config     JWTConfig
		reqURL     string // "/" if empty
		hdrAuth    string
		hdrCookie  string // test.Request doesn't provide SetCookie(); use name=val
		formValues map[string]string
	}{
		{
			name:     "No signing key provided",
			expPanic: true,
		},
		{
			name:       "Unexpected signing method",
			expErrCode: http.StatusBadRequest,
			config: JWTConfig{
				SigningKey:    validKey,
				SigningMethod: "RS256",
			},
		},
		{
			name:       "Invalid key",
			expErrCode: http.StatusUnauthorized,
			hdrAuth:    validAuth,
			config:     JWTConfig{SigningKey: invalidKey},
		},
		{
			name:    "Valid JWT",
			hdrAuth: validAuth,
			config:  JWTConfig{SigningKey: validKey},
		},
		{
			name:    "Valid JWT with custom AuthScheme",
			hdrAuth: "Token" + " " + token,
			config:  JWTConfig{AuthScheme: "Token", SigningKey: validKey},
		},
		{
			name:    "Valid JWT with custom claims",
			hdrAuth: validAuth,
			config: JWTConfig{
				Claims:     &jwtCustomClaims{},
				SigningKey: []byte("secret"),
			},
		},
		{
			name:       "Invalid Authorization header",
			hdrAuth:    "invalid-auth",
			expErrCode: http.StatusBadRequest,
			config:     JWTConfig{SigningKey: validKey},
		},
		{
			name:       "Empty header auth field",
			config:     JWTConfig{SigningKey: validKey},
			expErrCode: http.StatusBadRequest,
		},
		{
			name: "Valid query method",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL: "/?a=b&jwt=" + token,
		},
		{
			name: "Invalid query param name",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL:     "/?a=b&jwtxyz=" + token,
			expErrCode: http.StatusBadRequest,
		},
		{
			name: "Invalid query param value",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL:     "/?a=b&jwt=invalid-token",
			expErrCode: http.StatusUnauthorized,
		},
		{
			name: "Empty query",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL:     "/?a=b",
			expErrCode: http.StatusBadRequest,
		},
		{
			name: "Valid param method",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "param:jwt",
			},
			reqURL: "/" + token,
		},
		{
			name: "Valid cookie method",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "cookie:jwt",
			},
			hdrCookie: "jwt=" + token,
		},
		{
			name: "Multiple jwt lookuop",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt,cookie:jwt",
			},
			hdrCookie: "jwt=" + token,
		},
		{
			name: "Invalid token with cookie method",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "cookie:jwt",
			},
			expErrCode: http.StatusUnauthorized,
			hdrCookie:  "jwt=invalid",
		},
		{
			name: "Empty cookie",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "cookie:jwt",
			},
			expErrCode: http.StatusBadRequest,
		},
		{
			name: "Valid form method",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "form:jwt",
			},
			formValues: map[string]string{"jwt": token},
		},
		{
			name: "Invalid token with form method",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "form:jwt",
			},
			expErrCode: http.StatusUnauthorized,
			formValues: map[string]string{"jwt": "invalid"},
		},
		{
			name: "Empty form field",
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "form:jwt",
			},
			expErrCode: http.StatusBadRequest,
		},
		{
			name:    "Valid JWT with a valid key using a user-defined KeyFunc",
			hdrAuth: validAuth,
			config: JWTConfig{
				KeyFunc: func(*jwt.Token) (interface{}, error) {
					return validKey, nil
				},
			},
		},
		{
			name:    "Valid JWT with an invalid key using a user-defined KeyFunc",
			hdrAuth: validAuth,
			config: JWTConfig{
				KeyFunc: func(*jwt.Token) (interface{}, error) {
					return invalidKey, nil
				},
			},
			expErrCode: http.StatusUnauthorized,
		},
		{
			name:    "Token verification does not pass using a user-defined KeyFunc",
			hdrAuth: validAuth,
			config: JWTConfig{
				KeyFunc: func(*jwt.Token) (interface{}, error) {
					return nil, errors.New("faulty KeyFunc")
				},
			},
			expErrCode: http.StatusUnauthorized,
		},
		{
			name:    "Valid JWT with lower case AuthScheme",
			hdrAuth: strings.ToLower(DefaultJWTConfig.AuthScheme) + " " + token,
			config:  JWTConfig{SigningKey: validKey},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
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
				c.SetParamNames("jwt")
				c.SetParamValues(token)
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
				assert.Equal(t, tc.expErrCode, he.Code, tc.name)
				return
			}

			h := JWTWithConfig(tc.config)(handler)
			if assert.NoError(t, h(c), tc.name) {
				user := c.Get("user").(*jwt.Token)
				switch claims := user.Claims.(type) {
				case jwt.MapClaims:
					assert.Equal(t, claims["name"], "John Doe", tc.name)
				case *jwtCustomClaims:
					assert.Equal(t, claims.Name, "John Doe", tc.name)
					assert.Equal(t, claims.Admin, true, tc.name)
				default:
					panic("unexpected type of claims")
				}
			}
		})
	}
}

func TestJWTwithKID(t *testing.T) {
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	firstToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiIsImtpZCI6ImZpcnN0T25lIn0.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.w5VGpHOe0jlNgf7jMVLHzIYH_XULmpUlreJnilwSkWk"
	secondToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiIsImtpZCI6InNlY29uZE9uZSJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.sdghDYQ85jdh0hgQ6bKbMguLI_NSPYWjkhVJkee-yZM"
	wrongToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiIsImtpZCI6InNlY29uZE9uZSJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.RyhLybtVLpoewF6nz9YN79oXo32kAtgUxp8FNwTkb90"
	staticToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.1_-XFYUPpJfgsaGwYhgZEt7hfySMg-a3GN-nfZmbW7o"
	validKeys := map[string]interface{}{"firstOne": []byte("first_secret"), "secondOne": []byte("second_secret")}
	invalidKeys := map[string]interface{}{"thirdOne": []byte("third_secret")}
	staticSecret := []byte("static_secret")
	invalidStaticSecret := []byte("invalid_secret")

	for _, tc := range []struct {
		expErrCode int // 0 for Success
		config     JWTConfig
		hdrAuth    string
		info       string
	}{
		{
			hdrAuth: DefaultJWTConfig.AuthScheme + " " + firstToken,
			config:  JWTConfig{SigningKeys: validKeys},
			info:    "First token valid",
		},
		{
			hdrAuth: DefaultJWTConfig.AuthScheme + " " + secondToken,
			config:  JWTConfig{SigningKeys: validKeys},
			info:    "Second token valid",
		},
		{
			expErrCode: http.StatusUnauthorized,
			hdrAuth:    DefaultJWTConfig.AuthScheme + " " + wrongToken,
			config:     JWTConfig{SigningKeys: validKeys},
			info:       "Wrong key id token",
		},
		{
			hdrAuth: DefaultJWTConfig.AuthScheme + " " + staticToken,
			config:  JWTConfig{SigningKey: staticSecret},
			info:    "Valid static secret token",
		},
		{
			expErrCode: http.StatusUnauthorized,
			hdrAuth:    DefaultJWTConfig.AuthScheme + " " + staticToken,
			config:     JWTConfig{SigningKey: invalidStaticSecret},
			info:       "Invalid static secret",
		},
		{
			expErrCode: http.StatusUnauthorized,
			hdrAuth:    DefaultJWTConfig.AuthScheme + " " + firstToken,
			config:     JWTConfig{SigningKeys: invalidKeys},
			info:       "Invalid keys first token",
		},
		{
			expErrCode: http.StatusUnauthorized,
			hdrAuth:    DefaultJWTConfig.AuthScheme + " " + secondToken,
			config:     JWTConfig{SigningKeys: invalidKeys},
			info:       "Invalid keys second token",
		},
	} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderAuthorization, tc.hdrAuth)
		c := e.NewContext(req, res)

		if tc.expErrCode != 0 {
			h := JWTWithConfig(tc.config)(handler)
			he := h(c).(*echo.HTTPError)
			assert.Equal(t, tc.expErrCode, he.Code, tc.info)
			continue
		}

		h := JWTWithConfig(tc.config)(handler)
		if assert.NoError(t, h(c), tc.info) {
			user := c.Get("user").(*jwt.Token)
			switch claims := user.Claims.(type) {
			case jwt.MapClaims:
				assert.Equal(t, claims["name"], "John Doe", tc.info)
			case *jwtCustomClaims:
				assert.Equal(t, claims.Name, "John Doe", tc.info)
				assert.Equal(t, claims.Admin, true, tc.info)
			default:
				panic("unexpected type of claims")
			}
		}
	}
}

func TestJWTConfig_skipper(t *testing.T) {
	e := echo.New()

	e.Use(JWTWithConfig(JWTConfig{
		Skipper: func(context echo.Context) bool {
			return true // skip everything
		},
		SigningKey: []byte("secret"),
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
		SigningKey: []byte("secret"),
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, DefaultJWTConfig.AuthScheme+" eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
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
				SigningKey: []byte("secret"),
				ErrorHandler: func(err error) error {
					return echo.NewHTTPError(http.StatusTeapot, "custom_error")
				},
			},
			expectStatusCode: http.StatusTeapot,
		},
		{
			name: "ok, ErrorHandlerWithContext is executed",
			given: JWTConfig{
				SigningKey: []byte("secret"),
				ErrorHandlerWithContext: func(err error, context echo.Context) error {
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
				SigningKey: []byte("secret"),
				ErrorHandler: func(err error) error {
					return echo.NewHTTPError(http.StatusTeapot, "ErrorHandler: "+err.Error())
				},
			},
			expectErr: "{\"message\":\"ErrorHandler: parsing failed\"}\n",
		},
		{
			name: "ok, ErrorHandlerWithContext is executed",
			given: JWTConfig{
				SigningKey: []byte("secret"),
				ErrorHandlerWithContext: func(err error, context echo.Context) error {
					return echo.NewHTTPError(http.StatusTeapot, "ErrorHandlerWithContext: "+err.Error())
				},
			},
			expectErr: "{\"message\":\"ErrorHandlerWithContext: parsing failed\"}\n",
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
			config.ParseTokenFunc = func(auth string, c echo.Context) (interface{}, error) {
				parseTokenCalled = true
				return nil, errors.New("parsing failed")
			}
			e.Use(JWTWithConfig(config))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAuthorization, DefaultJWTConfig.AuthScheme+" eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
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
		ParseTokenFunc: func(auth string, c echo.Context) (interface{}, error) {
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
	req.Header.Set(echo.HeaderAuthorization, DefaultJWTConfig.AuthScheme+" eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusTeapot, res.Code)
}

func TestJWTConfig_TokenLookupFuncs(t *testing.T) {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)
		return c.JSON(http.StatusOK, token.Claims)
	})

	e.Use(JWTWithConfig(JWTConfig{
		TokenLookupFuncs: []ValuesExtractor{
			func(c echo.Context) ([]string, error) {
				return []string{c.Request().Header.Get("X-API-Key")}, nil
			},
		},
		SigningKey: []byte("secret"),
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"admin":true,"name":"John Doe","sub":"1234567890"}`+"\n", res.Body.String())
}

func TestJWTConfig_SuccessHandler(t *testing.T) {
	var testCases = []struct {
		name         string
		givenToken   string
		expectCalled bool
		expectStatus int
	}{
		{
			name:         "ok, success handler is called",
			givenToken:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ",
			expectCalled: true,
			expectStatus: http.StatusOK,
		},
		{
			name:         "nok, success handler is not called",
			givenToken:   "x.x.x",
			expectCalled: false,
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			e.GET("/", func(c echo.Context) error {
				token := c.Get("user").(*jwt.Token)
				return c.JSON(http.StatusOK, token.Claims)
			})

			wasCalled := false
			e.Use(JWTWithConfig(JWTConfig{
				SuccessHandler: func(c echo.Context) {
					wasCalled = true
				},
				SigningKey: []byte("secret"),
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderAuthorization, "bearer "+tc.givenToken)
			res := httptest.NewRecorder()

			e.ServeHTTP(res, req)

			assert.Equal(t, tc.expectCalled, wasCalled)
			assert.Equal(t, tc.expectStatus, res.Code)
		})
	}
}

func TestJWTConfig_ContinueOnIgnoredError(t *testing.T) {
	var testCases = []struct {
		name                       string
		whenContinueOnIgnoredError bool
		givenToken                 string
		expectStatus               int
		expectBody                 string
	}{
		{
			name:                       "no error handler is called",
			whenContinueOnIgnoredError: true,
			givenToken:                 "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ",
			expectStatus:               http.StatusTeapot,
			expectBody:                 "",
		},
		{
			name:                       "ContinueOnIgnoredError is false and error handler is called for missing token",
			whenContinueOnIgnoredError: false,
			givenToken:                 "",
			// empty response with 200. This emulates previous behaviour when error handler swallowed the error
			expectStatus: http.StatusOK,
			expectBody:   "",
		},
		{
			name:                       "error handler is called for missing token",
			whenContinueOnIgnoredError: true,
			givenToken:                 "",
			expectStatus:               http.StatusTeapot,
			expectBody:                 "public-token",
		},
		{
			name:                       "error handler is called for invalid token",
			whenContinueOnIgnoredError: true,
			givenToken:                 "x.x.x",
			expectStatus:               http.StatusUnauthorized,
			expectBody:                 "{\"message\":\"Unauthorized\"}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			e.GET("/", func(c echo.Context) error {
				testValue, _ := c.Get("test").(string)
				return c.String(http.StatusTeapot, testValue)
			})

			e.Use(JWTWithConfig(JWTConfig{
				ContinueOnIgnoredError: tc.whenContinueOnIgnoredError,
				SigningKey:             []byte("secret"),
				ErrorHandlerWithContext: func(err error, c echo.Context) error {
					if err == ErrJWTMissing {
						c.Set("test", "public-token")
						return nil
					}
					return echo.ErrUnauthorized
				},
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.givenToken != "" {
				req.Header.Set(echo.HeaderAuthorization, "bearer "+tc.givenToken)
			}
			res := httptest.NewRecorder()

			e.ServeHTTP(res, req)

			assert.Equal(t, tc.expectStatus, res.Code)
			assert.Equal(t, tc.expectBody, res.Body.String())
		})
	}
}
