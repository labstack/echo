package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	e := echo.New()
	buf := new(bytes.Buffer)
	e.Logger = &testLogger{output: buf}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Recover()(func(c echo.Context) error {
		panic("test")
	})
	err := h(c)
	assert.Contains(t, err.Error(), "[PANIC RECOVER] test goroutine")
	assert.Equal(t, http.StatusOK, rec.Code) // status is still untouched. err is returned from middleware chain
	assert.Contains(t, buf.String(), "")     // nothing is logged
}

func TestRecover_skipper(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	config := RecoverConfig{
		Skipper: func(c echo.Context) bool {
			return true
		},
	}
	h := RecoverWithConfig(config)(func(c echo.Context) error {
		panic("testPANIC")
	})

	var err error
	assert.Panics(t, func() {
		err = h(c)
	})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code) // status is still untouched. err is returned from middleware chain
}

func TestRecoverErrAbortHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Recover()(func(c echo.Context) error {
		panic(http.ErrAbortHandler)
	})
	defer func() {
		r := recover()
		if r == nil {
			assert.Fail(t, "expecting `http.ErrAbortHandler`, got `nil`")
		} else {
			if err, ok := r.(error); ok {
				assert.ErrorIs(t, err, http.ErrAbortHandler)
			} else {
				assert.Fail(t, "not of error type")
			}
		}
	}()

	hErr := h(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, hErr.Error(), "PANIC RECOVER")
}

func TestRecoverWithConfig(t *testing.T) {
	var testCases = []struct {
		name             string
		givenNoPanic     bool
		whenConfig       RecoverConfig
		expectErrContain string
		expectErr        string
	}{
		{
			name:             "ok, default config",
			whenConfig:       DefaultRecoverConfig,
			expectErrContain: "[PANIC RECOVER] testPANIC goroutine",
		},
		{
			name:             "ok, no panic",
			givenNoPanic:     true,
			whenConfig:       DefaultRecoverConfig,
			expectErrContain: "",
		},
		{
			name: "ok, DisablePrintStack",
			whenConfig: RecoverConfig{
				DisablePrintStack: true,
			},
			expectErr: "testPANIC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			config := tc.whenConfig
			h := RecoverWithConfig(config)(func(c echo.Context) error {
				if tc.givenNoPanic {
					return nil
				}
				panic("testPANIC")
			})

			err := h(c)

			if tc.expectErrContain != "" {
				assert.Contains(t, err.Error(), tc.expectErrContain)
			} else if tc.expectErr != "" {
				assert.Contains(t, err.Error(), tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, http.StatusOK, rec.Code) // status is still untouched. err is returned from middleware chain
		})
	}
}
