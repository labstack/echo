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
		TokenLookup:  "query:csrf;form:csrf;header:" + echo.HeaderXCSRFToken,
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
	req.Header().Set(echo.HeaderXCSRFToken, "foo")
	he := h(c).(*echo.HTTPError)
	assert.Equal(t, http.StatusForbidden, he.Code)

	// Valid CSRF token from header
	salt, _ := generateSalt(8)
	token := generateCSRFToken([]byte("secret"), salt)
	req.Header().Set(echo.HeaderXCSRFToken, token)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Status())
	}

	// Valid CSRF token from query
	req = test.NewRequest(echo.POST, "/?csrf="+token, nil)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Status())
	}

	// Valid CSRF token from form
	body := strings.NewReader(url.Values{"csrf": {token}}.Encode())
	req = test.NewRequest(echo.POST, "/", body)
	req.Header().Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
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
	token := csrfTokenFromForm("csrf")(c)
	assert.Equal(t, "token", token)
	token = csrfTokenFromForm("invalid")(c)
	assert.Equal(t, token, "")
}

func TestCSRFTokenFromQuery(t *testing.T) {
	q := make(url.Values)
	q.Set("csrf", "token")
	e := echo.New()
	req := test.NewRequest(echo.GET, "/?"+q.Encode(), nil)
	req.Header().Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	c := e.NewContext(req, nil)
	token := csrfTokenFromQuery("csrf")(c)
	assert.Equal(t, "token", token)
	token = csrfTokenFromQuery("invalid")(c)
	assert.Equal(t, token, "")
	csrfTokenFromQuery("csrf")
}
