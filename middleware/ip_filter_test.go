package middleware

import (
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIpFilterWithConfig(t *testing.T) {
	a := assert.New(t)

	testTable := []struct{
		Config IpFilterConfig
		Ip string
		Expect error
	}{
		{
			Config: IpFilterConfig{
				IpFilters: []IpFilterFunc{
					func (ip string) bool {
						if ip[0] == '1' {
							return true
						}
						return false
					},
				},
			},
			Ip: "123.123.123.123",
			// Blocked by IpFilters.
			Expect: echo.ErrForbidden,
		},
		{
			Config: IpFilterConfig{
				IpFilters: []IpFilterFunc{
					func (ip string) bool {
						if ip[0] == '1' {
							return true
						}
						return false
					},
				},
				WhiteList: []string{
					"123.123.123.123",
				},
			},
			Ip: "223.123.123.123",
			// Blocked by WhiteList.
			Expect: echo.ErrForbidden,
		},
		{
			Config: IpFilterConfig{
				IpFilters: []IpFilterFunc{
					func (ip string) bool {
						if ip[0] == '1' {
							return true
						}
						return false
					},
				},
				WhiteList: []string{
					"223.123.123.123",
				},
			},
			Ip: "223.123.123.123",
			// Not blocked.
			Expect: nil,
		},
		{
			Config: IpFilterConfig{
				IpFilters: []IpFilterFunc{
					func (ip string) bool {
						if ip[0] == '1' {
							return true
						}
						return false
					},
				},
				WhiteList: []string{
					"223.123.123.0/24",
				},
			},
			Ip: "223.123.123.230",
			// Not blocked.
			Expect: nil,
		},
		{
			Config: IpFilterConfig{
				IpFilters: []IpFilterFunc{
					func (ip string) bool {
						if ip[0] == '1' {
							return true
						}
						return false
					},
				},
				WhiteList: []string{
					"223.123.123.0/24",
				},
				BlackList: []string{
					"223.123.123.230",
				},
			},
			Ip: "223.123.123.230",
			// Blocked by BlackList.
			Expect: echo.ErrForbidden,
		},
		{
			// It will be DefaultIpFilterConfig.
			Config: IpFilterConfig{},
			// Blocked by WhiteList.
			Ip: "123.123.123.123",
			Expect: echo.ErrForbidden,
		},
	}

	e := echo.New()

	for idx, testCase := range testTable {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = testCase.Ip
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)
		h := IpFilterWithConfig(testCase.Config)(func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		switch err := h(c); err.(type) {
		case nil:
			a.EqualValues(testCase.Expect, err, "testTable[%d]", idx)
		case *echo.HTTPError:
			a.EqualValues(testCase.Expect, err, "testTable[%d]", idx)
		}
	}
}

func TestIpFilter(t *testing.T) {
	a := assert.New(t)

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "123.123.123.123"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := IpFilter()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// Blocked by WhiteList.
	a.Error(h(c))
}
