package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestSecure(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Default
	err := Secure()(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, "1; mode=block", rec.Header().Get(echo.HeaderXXSSProtection))
	assert.Equal(t, "nosniff", rec.Header().Get(echo.HeaderXContentTypeOptions))
	assert.Equal(t, "SAMEORIGIN", rec.Header().Get(echo.HeaderXFrameOptions))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderStrictTransportSecurity))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderContentSecurityPolicy))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderReferrerPolicy))
}

func TestSecureWithConfig(t *testing.T) {
	e := echo.New()
	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mw, err := SecureConfig{
		XSSProtection:         "",
		ContentTypeNosniff:    "",
		XFrameOptions:         "",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "origin",
	}.ToMiddleware()
	assert.NoError(t, err)

	err = mw(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, "", rec.Header().Get(echo.HeaderXXSSProtection))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderXContentTypeOptions))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderXFrameOptions))
	assert.Equal(t, "max-age=3600; includeSubdomains", rec.Header().Get(echo.HeaderStrictTransportSecurity))
	assert.Equal(t, "default-src 'self'", rec.Header().Get(echo.HeaderContentSecurityPolicy))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderContentSecurityPolicyReportOnly))
	assert.Equal(t, "origin", rec.Header().Get(echo.HeaderReferrerPolicy))

}

func TestSecureWithConfig_CSPReportOnly(t *testing.T) {
	// Custom with CSPReportOnly flag
	e := echo.New()
	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SecureWithConfig(SecureConfig{
		XSSProtection:         "",
		ContentTypeNosniff:    "",
		XFrameOptions:         "",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
		CSPReportOnly:         true,
		ReferrerPolicy:        "origin",
	})(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, "", rec.Header().Get(echo.HeaderXXSSProtection))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderXContentTypeOptions))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderXFrameOptions))
	assert.Equal(t, "max-age=3600; includeSubdomains", rec.Header().Get(echo.HeaderStrictTransportSecurity))
	assert.Equal(t, "default-src 'self'", rec.Header().Get(echo.HeaderContentSecurityPolicyReportOnly))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderContentSecurityPolicy))
	assert.Equal(t, "origin", rec.Header().Get(echo.HeaderReferrerPolicy))
}

func TestSecureWithConfig_HSTSPreloadEnabled(t *testing.T) {
	// Custom with CSPReportOnly flag
	e := echo.New()
	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Custom, with preload option enabled
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SecureWithConfig(SecureConfig{
		HSTSMaxAge:         3600,
		HSTSPreloadEnabled: true,
	})(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, "max-age=3600; includeSubdomains; preload", rec.Header().Get(echo.HeaderStrictTransportSecurity))

}

func TestSecureWithConfig_HSTSExcludeSubdomains(t *testing.T) {
	// Custom with CSPReportOnly flag
	e := echo.New()
	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Custom, with preload option enabled and subdomains excluded
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SecureWithConfig(SecureConfig{
		HSTSMaxAge:            3600,
		HSTSPreloadEnabled:    true,
		HSTSExcludeSubdomains: true,
	})(h)(c)
	assert.NoError(t, err)

	assert.Equal(t, "max-age=3600; preload", rec.Header().Get(echo.HeaderStrictTransportSecurity))
}
