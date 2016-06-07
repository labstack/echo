package middleware

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestCSRF(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	csrf := CSRFWithConfig(CSRFConfig{
		Secret:       []byte("secret"),
		CookiePath:   "/",
		CookieDomain: "labstack.com",
	})
	h := csrf(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// No secret
	assert.Panics(t, func() {
		CSRF(nil)
	})

	// Generate CSRF token
	h(c)
	assert.Contains(t, rec.Header().Get(echo.HeaderSetCookie), "csrf")

	// Empty/invalid CSRF token
	req = test.NewRequest(echo.POST, "/", nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	req.Header().Set(echo.HeaderXCSRFToken, "")
	he := h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusForbidden, he.Code)

	// Valid CSRF token
	salt, _ := generateSalt(8)
	token := generateCSRFToken([]byte("secret"), salt)
	req.Header().Set(echo.HeaderXCSRFToken, token)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Status())
	}
}

func TestCSRFTokenFromForm(t *testing.T) {
	f := make(url.Values)
	f.Set("csrf", "token")
	e := echo.New()
	req := test.NewRequest(echo.POST, "/", strings.NewReader(f.Encode()))
	req.Header().Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	c := e.NewContext(req, nil)
	token, err := csrfTokenFromForm("csrf")(c)
	if assert.NoError(t, err) {
		assert.Equal(t, "token", token)
	}
	token, err = csrfTokenFromForm("invalid")(c)
	assert.Error(t, err)
}

func TestCSRFTokenFromQuery(t *testing.T) {
	q := make(url.Values)
	q.Set("csrf", "token")
	e := echo.New()
	req := test.NewRequest(echo.GET, "/?"+q.Encode(), nil)
	req.Header().Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	c := e.NewContext(req, nil)
	token, err := csrfTokenFromQuery("csrf")(c)
	if assert.NoError(t, err) {
		assert.Equal(t, "token", token)
	}
	token, err = csrfTokenFromQuery("invalid")(c)
	assert.Error(t, err)
	csrfTokenFromQuery("csrf")
}
