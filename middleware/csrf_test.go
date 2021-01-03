package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"github.com/stretchr/testify/assert"
)

func TestCSRF(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	csrf := CSRFWithConfig(CSRFConfig{
		TokenLength: 16,
	})
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
	token := random.String(16)
	req.Header.Set(echo.HeaderCookie, "_csrf="+token)
	req.Header.Set(echo.HeaderXCSRFToken, token)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestCSRFTokenFromForm(t *testing.T) {
	f := make(url.Values)
	f.Set("csrf", "token")
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	c := e.NewContext(req, nil)
	token, err := csrfTokenFromForm("csrf")(c)
	if assert.NoError(t, err) {
		assert.Equal(t, "token", token)
	}
	_, err = csrfTokenFromForm("invalid")(c)
	assert.Error(t, err)
}

func TestCSRFTokenFromQuery(t *testing.T) {
	q := make(url.Values)
	q.Set("csrf", "token")
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.URL.RawQuery = q.Encode()
	c := e.NewContext(req, nil)
	token, err := csrfTokenFromQuery("csrf")(c)
	if assert.NoError(t, err) {
		assert.Equal(t, "token", token)
	}
	_, err = csrfTokenFromQuery("invalid")(c)
	assert.Error(t, err)
	csrfTokenFromQuery("csrf")
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
	fmt.Println(rec.Header()["Set-Cookie"])
	assert.NotRegexp(t, "SameSite=", rec.Header()["Set-Cookie"])
}
