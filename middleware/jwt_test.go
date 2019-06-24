package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
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

func TestJWT(t *testing.T) {
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
	validKey := []byte("secret")
	invalidKey := []byte("invalid-key")
	validAuth := DefaultJWTConfig.AuthScheme + " " + token

	for _, tc := range []struct {
		expPanic   bool
		expErrCode int // 0 for Success
		config     JWTConfig
		reqURL     string // "/" if empty
		hdrAuth    string
		hdrCookie  string // test.Request doesn't provide SetCookie(); use name=val
		info       string
	}{
		{
			expPanic: true,
			info:     "No signing key provided",
		},
		{
			expErrCode: http.StatusBadRequest,
			config: JWTConfig{
				SigningKey:    validKey,
				SigningMethod: "RS256",
			},
			info: "Unexpected signing method",
		},
		{
			expErrCode: http.StatusUnauthorized,
			hdrAuth:    validAuth,
			config:     JWTConfig{SigningKey: invalidKey},
			info:       "Invalid key",
		},
		{
			hdrAuth: validAuth,
			config:  JWTConfig{SigningKey: validKey},
			info:    "Valid JWT",
		},
		{
			hdrAuth: "Token" + " " + token,
			config:  JWTConfig{AuthScheme: "Token", SigningKey: validKey},
			info:    "Valid JWT with custom AuthScheme",
		},
		{
			hdrAuth: validAuth,
			config: JWTConfig{
				Claims:     &jwtCustomClaims{},
				SigningKey: []byte("secret"),
			},
			info: "Valid JWT with custom claims",
		},
		{
			hdrAuth:    "invalid-auth",
			expErrCode: http.StatusBadRequest,
			config:     JWTConfig{SigningKey: validKey},
			info:       "Invalid Authorization header",
		},
		{
			config:     JWTConfig{SigningKey: validKey},
			expErrCode: http.StatusBadRequest,
			info:       "Empty header auth field",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL: "/?a=b&jwt=" + token,
			info:   "Valid query method",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL:     "/?a=b&jwtxyz=" + token,
			expErrCode: http.StatusBadRequest,
			info:       "Invalid query param name",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL:     "/?a=b&jwt=invalid-token",
			expErrCode: http.StatusUnauthorized,
			info:       "Invalid query param value",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL:     "/?a=b",
			expErrCode: http.StatusBadRequest,
			info:       "Empty query",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "param:jwt",
			},
			reqURL: "/" + token,
			info:   "Valid param method",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "cookie:jwt",
			},
			hdrCookie: "jwt=" + token,
			info:      "Valid cookie method",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "cookie:jwt",
			},
			expErrCode: http.StatusUnauthorized,
			hdrCookie:  "jwt=invalid",
			info:       "Invalid token with cookie method",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "cookie:jwt",
			},
			expErrCode: http.StatusBadRequest,
			info:       "Empty cookie",
		},
	} {
		if tc.reqURL == "" {
			tc.reqURL = "/"
		}

		req := httptest.NewRequest(http.MethodGet, tc.reqURL, nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderAuthorization, tc.hdrAuth)
		req.Header.Set(echo.HeaderCookie, tc.hdrCookie)
		c := e.NewContext(req, res)

		if tc.reqURL == "/" + token {
			c.SetParamNames("jwt")
			c.SetParamValues(token)
		}

		if tc.expPanic {
			assert.Panics(t, func() {
				JWTWithConfig(tc.config)
			}, tc.info)
			continue
		}

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

func TestJWTwithKID(t *testing.T) {
	test := assert.New(t)

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
			test.Equal(tc.expErrCode, he.Code, tc.info)
			continue
		}

		h := JWTWithConfig(tc.config)(handler)
		if test.NoError(h(c), tc.info) {
			user := c.Get("user").(*jwt.Token)
			switch claims := user.Claims.(type) {
			case jwt.MapClaims:
				test.Equal(claims["name"], "John Doe", tc.info)
			case *jwtCustomClaims:
				test.Equal(claims.Name, "John Doe", tc.info)
				test.Equal(claims.Admin, true, tc.info)
			default:
				panic("unexpected type of claims")
			}
		}
	}
}
