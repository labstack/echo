package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

type (
	// IpFilterConfig defines the config for IpFilter middleware.
	IpFilterConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// IpFilters is FilterFunc list that filter with custom logic.
		IpFilters []IpFilterFunc

		// WhiteList is an allowed ip list.
		WhiteList []string

		// BlackList is a disallowed ip list.
		BlackList []string
	}

	// FilterFunc defines a function to filter ip with custom logic.
	// If FilterFunc returns true, that ip will be blocked.
	IpFilterFunc func(string) bool
)

var (
	// DefaultIpFilterConfig is the default IpFilter middleware config.
	DefaultIpFilterConfig = IpFilterConfig{
		Skipper: DefaultSkipper,
		WhiteList: []string{
			"0.0.0.0/32",
		},
	}
)

// IpFilter returns a IpFilter middleware.
//
// IpFilter middleware filters requests by ip matching.
// IpFilter filters ip with `IpFilters []IpFilterFunc` field first.
// And then, It checks `WhiteList []string` and `BlackList []string` to filtering requests.
func IpFilter() echo.MiddlewareFunc {
	return IpFilterWithConfig(DefaultIpFilterConfig)
}

// IpFilterWithConfig returns a IpFilter middleware with config.
// See: `IpFilter()`.
func IpFilterWithConfig(config IpFilterConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultIpFilterConfig.Skipper
	}

	whiteList := make([]*net.IPNet, 0, len(config.WhiteList))
	for _, allowed := range config.WhiteList {
		if strings.ContainsRune(allowed, '/') {
			_, ipNet, err := net.ParseCIDR(allowed)
			if err != nil {
				panic(err)
			}

			whiteList = append(whiteList, ipNet)
		} else {
			ipNet := &net.IPNet{
				IP: net.ParseIP(allowed),
			}
			if ipNet.IP == nil {
				panic((&net.ParseError{Type: "IP address", Text: allowed}).Error())
			}

			switch len(ipNet.IP) {
			case net.IPv4len:
				ipNet.Mask = net.CIDRMask(32, 32)
			case net.IPv6len:
				ipNet.Mask = net.CIDRMask(128, 128)
			default:
				panic((&net.ParseError{Type: "IP address", Text: allowed}).Error())
			}

			whiteList = append(whiteList, ipNet)
		}
	}

	blackList := make([]*net.IPNet, 0, len(config.BlackList))
	for _, disallowed := range config.BlackList {
		if strings.ContainsRune(disallowed, '/') {
			_, ipNet, err := net.ParseCIDR(disallowed)
			if err != nil {
				panic(err)
			}

			blackList = append(blackList, ipNet)
		} else {
			ipNet := &net.IPNet{
				IP: net.ParseIP(disallowed),
			}
			if ipNet.IP == nil {
				panic((&net.ParseError{Type: "IP address", Text: disallowed}).Error())
			}

			switch len(ipNet.IP) {
			case net.IPv4len:
				ipNet.Mask = net.CIDRMask(32, 32)
			case net.IPv6len:
				ipNet.Mask = net.CIDRMask(128, 128)
			default:
				panic((&net.ParseError{Type: "IP address", Text: disallowed}).Error())
			}

			blackList = append(blackList, ipNet)
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			ip, _, err := net.SplitHostPort(c.Request().RemoteAddr)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err)
			}

			for _, filter := range config.IpFilters {
				if filter(ip) {
					return echo.ErrForbidden
				}
			}

			parsedIp := net.ParseIP(ip)

			for _, allowed := range whiteList {
				if allowed.Contains(parsedIp) {

					for _, disallowed := range blackList {
						if disallowed.Contains(parsedIp) {
							return echo.ErrForbidden
						}
					}

					return next(c)
				}
			}

			return echo.ErrForbidden
		}
	}
}
