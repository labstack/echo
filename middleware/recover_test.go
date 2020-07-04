package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	e := echo.New()
	buf := new(bytes.Buffer)
	e.Logger.SetOutput(buf)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Recover()(echo.HandlerFunc(func(c echo.Context) error {
		panic("test")
	}))
	h(c)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, buf.String(), "PANIC RECOVER")
}

func TestRecoverWithConfig_LogLevel(t *testing.T) {
	tests := []struct {
		logLevel  log.Lvl
		levelName string
	}{{
		logLevel:  log.DEBUG,
		levelName: "DEBUG",
	}, {
		logLevel:  log.INFO,
		levelName: "INFO",
	}, {
		logLevel:  log.WARN,
		levelName: "WARN",
	}, {
		logLevel:  log.ERROR,
		levelName: "ERROR",
	}, {
		logLevel:  log.OFF,
		levelName: "OFF",
	}}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.levelName, func(t *testing.T) {
			e := echo.New()
			e.Logger.SetLevel(log.DEBUG)

			buf := new(bytes.Buffer)
			e.Logger.SetOutput(buf)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			config := DefaultRecoverConfig
			config.LogLevel = tt.logLevel
			h := RecoverWithConfig(config)(echo.HandlerFunc(func(c echo.Context) error {
				panic("test")
			}))

			h(c)

			assert.Equal(t, http.StatusInternalServerError, rec.Code)

			output := buf.String()
			if tt.logLevel == log.OFF {
				assert.Empty(t, output)
			} else {
				assert.Contains(t, output, "PANIC RECOVER")
				assert.Contains(t, output, fmt.Sprintf(`"level":"%s"`, tt.levelName))
			}
		})
	}
}
