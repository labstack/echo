package middleware

import (
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

func TestSecure(t *testing.T) {
	e := echo.New()
	req := test.NewRequest(echo.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Default
	Secure()(h)(c)
	assert.Equal(t, "1; mode=block", rec.Header().Get(echo.HeaderXXSSProtection))
	assert.Equal(t, "nosniff", rec.Header().Get(echo.HeaderXContentTypeOptions))
	assert.Equal(t, "SAMEORIGIN", rec.Header().Get(echo.HeaderXFrameOptions))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderStrictTransportSecurity))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderContentSecurityPolicy))

	// Custom
	req.Header().Set(echo.HeaderXForwardedProto, "https")
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)
	SecureWithConfig(SecureConfig{
		XSSProtection:         "",
		ContentTypeNosniff:    "",
		XFrameOptions:         "",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	})(h)(c)
	assert.Equal(t, "", rec.Header().Get(echo.HeaderXXSSProtection))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderXContentTypeOptions))
	assert.Equal(t, "", rec.Header().Get(echo.HeaderXFrameOptions))
	assert.Equal(t, "max-age=3600; includeSubdomains", rec.Header().Get(echo.HeaderStrictTransportSecurity))
	assert.Equal(t, "default-src 'self'", rec.Header().Get(echo.HeaderContentSecurityPolicy))
}
