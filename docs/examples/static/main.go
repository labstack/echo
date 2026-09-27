// SPDX-License-Identifier: MIT

// This complete example is the source for the Static middleware documentation.
package main

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:                 "public",
		EnablePathUnescaping: false, // Keep encoded slashes encoded when route guards protect files.
	}))

	if err := e.Start(":1323"); err != nil {
		e.Logger.Error("server stopped", "error", err)
	}
}
