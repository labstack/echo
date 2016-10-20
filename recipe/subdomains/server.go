package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

type (
	Host struct {
		Echo *echo.Echo
	}
)

func main() {
	// Hosts
	hosts := make(map[string]*Host)

	//-----
	// API
	//-----

	api := echo.New()
	api.Use(middleware.Logger())
	api.Use(middleware.Recover())

	hosts["api.localhost:1323"] = &Host{api}

	api.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "API")
	})

	//------
	// Blog
	//------

	blog := echo.New()
	blog.Use(middleware.Logger())
	blog.Use(middleware.Recover())

	hosts["blog.localhost:1323"] = &Host{blog}

	blog.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Blog")
	})

	//---------
	// Website
	//---------

	site := echo.New()
	site.Use(middleware.Logger())
	site.Use(middleware.Recover())

	hosts["localhost:1323"] = &Host{site}

	site.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Website")
	})

	// Server
	e := echo.New()
	e.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		host := hosts[req.Host()]

		if host == nil {
			err = echo.ErrNotFound
		} else {
			host.Echo.ServeHTTP(req, res)
		}

		return
	})
	e.Run(standard.New(":1323"))
}
