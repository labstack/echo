package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBodyDump(t *testing.T) {
	e := echo.New()
	hw := "Hello, World!"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	}

	requestBody := ""
	responseBody := ""
	mw := BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		requestBody = string(reqBody)
		responseBody = string(resBody)
	})

	if assert.NoError(t, mw(h)(c)) {
		assert.Equal(t, requestBody, hw)
		assert.Equal(t, responseBody, hw)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, hw, rec.Body.String())
	}

	// Must set default skipper
	BodyDumpWithConfig(BodyDumpConfig{
		Skipper: nil,
		Handler: func(c echo.Context, reqBody, resBody []byte) {
			requestBody = string(reqBody)
			responseBody = string(resBody)
		},
	})
}

func TestBodyDumpFails(t *testing.T) {
	e := echo.New()
	hw := "Hello, World!"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(hw))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := func(c echo.Context) error {
		return errors.New("some error")
	}

	mw := BodyDump(func(c echo.Context, reqBody, resBody []byte) {})

	if !assert.Error(t, mw(h)(c)) {
		t.FailNow()
	}

	assert.Panics(t, func() {
		mw = BodyDumpWithConfig(BodyDumpConfig{
			Skipper: nil,
			Handler: nil,
		})
	})

	assert.NotPanics(t, func() {
		mw = BodyDumpWithConfig(BodyDumpConfig{
			Skipper: func(c echo.Context) bool {
				return true
			},
			Handler: func(c echo.Context, reqBody, resBody []byte) {
			},
		})

		if !assert.Error(t, mw(h)(c)) {
			t.FailNow()
		}
	})
}
