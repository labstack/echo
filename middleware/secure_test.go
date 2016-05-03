package middleware

// func TestSecureWithConfig(t *testing.T) {
// 	e := echo.New()
//
// 	config := SecureConfig{
// 		STSMaxAge:             100,
// 		STSIncludeSubdomains:  true,
// 		FrameDeny:             true,
// 		FrameOptionsValue:     "",
// 		ContentTypeNosniff:    true,
// 		XssProtected:          true,
// 		XssProtectionValue:    "",
// 		ContentSecurityPolicy: "default-src 'self'",
// 		DisableProdCheck:      true,
// 	}
// 	secure := SecureWithConfig(config)
// 	h := secure(func(c echo.Context) error {
// 		return c.String(http.StatusOK, "test")
// 	})
//
// 	rq := test.NewRequest(echo.GET, "/", nil)
// 	rc := test.NewResponseRecorder()
// 	c := e.NewContext(rq, rc)
// 	h(c)
//
// 	assert.Equal(t, "max-age=100; includeSubdomains", rc.Header().Get(stsHeader))
// 	assert.Equal(t, "DENY", rc.Header().Get(frameOptionsHeader))
// 	assert.Equal(t, "nosniff", rc.Header().Get(contentTypeHeader))
// 	assert.Equal(t, xssProtectionValue, rc.Header().Get(xssProtectionHeader))
// 	assert.Equal(t, "default-src 'self'", rc.Header().Get(cspHeader))
// }
