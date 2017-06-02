// Package middleware CasbinAuth provides handlers to enable ACL, RBAC, ABAC authorization support.
// Simple Usage:
// package main
//
// import (
// 	"github.com/casbin/casbin"
// 	"github.com/labstack/echo"
// 	"github.com/labstack/echo/middleware"
// )
//
// func main() {
// 	e := echo.New()
//
// 	// mediate the access for every request
// 	e.Use(middleware.CasbinAuth(casbin.NewEnforcer("casbin_auth_model.conf", "casbin_auth_policy.csv")))
//
// 	e.Logger.Fatal(e.Start(":1323"))
// }
//
// Advanced Usage:
//
//	func main(){
//		ce := casbin.NewEnforcer("casbin_auth_model.conf", "")
//		ce.AddRoleForUser("alice", "admin")
//		ce.AddPolicy(...)
//
//		e := echo.New()
//
//		echo.Use(middleware.CasbinAuth(ce))
//
//		e.Logger.Fatal(e.Start(":1323"))
//	}
package middleware

import (
	"github.com/casbin/casbin"
	"github.com/labstack/echo"
)

type (
	// CasbinAuthConfig defines the config for CasbinAuth middleware.
	CasbinAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper
		// Enforcer CasbinAuth main rule.
		// Required.
		Enforcer *casbin.Enforcer
	}
)

var (
	// DefaultCasbinAuthConfig is the default CasbinAuth middleware config.
	DefaultCasbinAuthConfig = CasbinAuthConfig{
		Skipper: DefaultSkipper,
	}
)

// CasbinAuth returns an CasbinAuth middleware.
//
// For valid credentials it calls the next handler.
// For missing or invalid credentials, it sends "401 - Unauthorized" response.
func CasbinAuth(ce *casbin.Enforcer) echo.MiddlewareFunc {
	c := DefaultCasbinAuthConfig
	c.Enforcer = ce
	return CasbinAuthWithConfig(c)
}

// CasbinAuthWithConfig returns an CasbinAuth middleware with config.
// See `CasbinAuth()`.
func CasbinAuthWithConfig(config CasbinAuthConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultCasbinAuthConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) || config.CheckPermission(c) {
				return next(c)
			}

			return echo.ErrForbidden
		}
	}
}

// GetUserName gets the user name from the request.
// Currently, only HTTP basic authentication is supported
func (a *CasbinAuthConfig) GetUserName(c echo.Context) string {
	username, _, _ := c.Request().BasicAuth()
	return username
}

// CheckPermission checks the user/method/path combination from the request.
// Returns true (permission granted) or false (permission forbidden)
func (a *CasbinAuthConfig) CheckPermission(c echo.Context) bool {
	user := a.GetUserName(c)
	method := c.Request().Method
	path := c.Request().URL.Path
	return a.Enforcer.Enforce(user, path, method)
}
