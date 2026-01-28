// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRetryMax_successFirstAttempt(t *testing.T) {
	e := echo.New()
	attempts := 0

	e.Use(RetryMaxWithConfig(RetryMaxConfig{
		MaxAttempts: 3,
		MinTimeout:  10 * time.Millisecond,
		MaxTimeout:  20 * time.Millisecond,
	}))

	e.GET("/ok", func(c echo.Context) error {
		attempts++
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, attempts, "Should only be called once when successful")
}

func TestRetryMax_retriesUntilSuccess(t *testing.T) {
	e := echo.New()
	attempts := 0

	e.Use(RetryMaxWithConfig(RetryMaxConfig{
		MaxAttempts: 5,
		MinTimeout:  1 * time.Millisecond,
		MaxTimeout:  5 * time.Millisecond,
	}))

	e.GET("/sometimes", func(c echo.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("fail")
		}
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/sometimes", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 3, attempts, "Should retry until success")
}

func TestRetryMax_allAttemptsFail(t *testing.T) {
	e := echo.New()
	attempts := 0

	e.Use(RetryMaxWithConfig(RetryMaxConfig{
		MaxAttempts: 4,
		MinTimeout:  1 * time.Millisecond,
		MaxTimeout:  1 * time.Millisecond,
	}))

	e.GET("/fail", func(c echo.Context) error {
		attempts++
		return errors.New("always fail")
	})

	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, 4, attempts, "Should hit max attempts")
}

func TestRetryMax_skipper(t *testing.T) {
	e := echo.New()
	attempts := 0

	e.Use(RetryMaxWithConfig(RetryMaxConfig{
		Skipper: func(c echo.Context) bool {
			return true
		},
		MaxAttempts: 3,
		MinTimeout:  1 * time.Millisecond,
		MaxTimeout:  1 * time.Millisecond,
	}))

	e.GET("/no-retry", func(c echo.Context) error {
		attempts++
		return errors.New("fail")
	})

	req := httptest.NewRequest(http.MethodGet, "/no-retry", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, 1, attempts, "Skipper should prevent retries")
}

func TestRetryMax_respectsMinMaxTimeout(t *testing.T) {
	e := echo.New()
	attempts := 0
	start := time.Now()

	e.Use(RetryMaxWithConfig(RetryMaxConfig{
		MaxAttempts: 3,
		MinTimeout:  10 * time.Millisecond,
		MaxTimeout:  10 * time.Millisecond,
	}))

	e.GET("/delays", func(c echo.Context) error {
		attempts++
		return errors.New("fail")
	})

	req := httptest.NewRequest(http.MethodGet, "/delays", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	elapsed := time.Since(start)
	minExpected := 2 * 10 * time.Millisecond
	assert.True(t, elapsed >= minExpected, "Expected at least %v elapsed, got %v", minExpected, elapsed)
	assert.Equal(t, 3, attempts)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
