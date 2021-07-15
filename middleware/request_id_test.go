package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	rid := RequestID()
	h := rid(handler)
	err := h(c)
	assert.NoError(t, err)
	assert.Len(t, rec.Header().Get(echo.HeaderXRequestID), 32)
}

func TestMustRequestIDWithConfig_skipper(t *testing.T) {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusTeapot, "test")
	})

	generatorCalled := false
	e.Use(RequestIDWithConfig(RequestIDConfig{
		Skipper: func(c echo.Context) bool {
			return true
		},
		Generator: func() string {
			generatorCalled = true
			return "customGenerator"
		},
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusTeapot, res.Code)
	assert.Equal(t, "test", res.Body.String())

	assert.Equal(t, res.Header().Get(echo.HeaderXRequestID), "")
	assert.False(t, generatorCalled)
}

func TestMustRequestIDWithConfig_customGenerator(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	rid := RequestIDWithConfig(RequestIDConfig{
		Generator: func() string { return "customGenerator" },
	})
	h := rid(handler)
	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, rec.Header().Get(echo.HeaderXRequestID), "customGenerator")
}

func TestMustRequestIDWithConfig_RequestIDHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	called := false
	rid := RequestIDWithConfig(RequestIDConfig{
		Generator: func() string { return "customGenerator" },
		RequestIDHandler: func(c echo.Context, s string) {
			called = true
		},
	})
	h := rid(handler)
	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, rec.Header().Get(echo.HeaderXRequestID), "customGenerator")
	assert.True(t, called)
}

func TestRequestIDWithConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	rid, err := RequestIDConfig{}.ToMiddleware()
	assert.NoError(t, err)
	h := rid(handler)
	h(c)
	assert.Len(t, rec.Header().Get(echo.HeaderXRequestID), 32)

	// Custom generator
	rid = RequestIDWithConfig(RequestIDConfig{
		Generator: func() string { return "customGenerator" },
	})
	h = rid(handler)
	h(c)
	assert.Equal(t, rec.Header().Get(echo.HeaderXRequestID), "customGenerator")
}

func TestRequestID_IDNotAltered(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add(echo.HeaderXRequestID, "<sample-request-id>")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	rid := RequestIDWithConfig(RequestIDConfig{})
	h := rid(handler)
	_ = h(c)
	assert.Equal(t, rec.Header().Get(echo.HeaderXRequestID), "<sample-request-id>")
}

func TestRequestIDConfigDifferentHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	rid := RequestIDWithConfig(RequestIDConfig{TargetHeader: echo.HeaderXCorrelationID})
	h := rid(handler)
	h(c)
	assert.Len(t, rec.Header().Get(echo.HeaderXCorrelationID), 32)

	// Custom generator and handler
	customID := "customGenerator"
	calledHandler := false
	rid = RequestIDWithConfig(RequestIDConfig{
		Generator:    func() string { return customID },
		TargetHeader: echo.HeaderXCorrelationID,
		RequestIDHandler: func(_ echo.Context, id string) {
			calledHandler = true
			assert.Equal(t, customID, id)
		},
	})
	h = rid(handler)
	h(c)
	assert.Equal(t, rec.Header().Get(echo.HeaderXCorrelationID), "customGenerator")
	assert.True(t, calledHandler)
}
