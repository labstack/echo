package middleware

import (
	"net/http"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
	validKey := []byte("secret")
	invalidKey := []byte("invalid-key")
	validAuth := bearer + " " + token

	for _, tc := range []struct {
		expPanic   bool
		expErrCode int // 0 for Success
		config     JWTConfig
		reqURL     string // "/" if empty
		hdrAuth    string
		info       string
	}{
		{expPanic: true, info: "No signing key provided"},
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
			hdrAuth:    "invalid-auth",
			expErrCode: http.StatusBadRequest,
			config:     JWTConfig{SigningKey: validKey},
			info:       "Invalid Authorization header",
		},
		{
			config:     JWTConfig{SigningKey: validKey},
			hdrAuth:    "",
			expErrCode: http.StatusBadRequest,
			info:       "Empty header auth field",
		},
	} {

		if tc.reqURL == "" {
			tc.reqURL = "/"
		}

		req := test.NewRequest(echo.GET, tc.reqURL, nil)
		res := test.NewResponseRecorder()
		req.Header().Set(echo.HeaderAuthorization, tc.hdrAuth)
		c := e.NewContext(req, res)

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
			claims := user.Claims.(jwt.MapClaims)
			assert.Equal(t, claims["name"], "John Doe", tc.info)
		}
	}
}
