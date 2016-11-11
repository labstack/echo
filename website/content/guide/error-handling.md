+++
title = "Error Handling"
description = "Error handling in Echo"
[menu.side]
  name = "Error Handling"
  parent = "guide"
  weight = 8
+++

## Error Handling

Echo advocates centralized HTTP error handling by returning error from middleware
or handlers.

- Log errors from a unified location
- Send customized HTTP responses

For example, when basic auth middleware finds invalid credentials it returns
`401 - Unauthorized` error, aborting the current HTTP request.

```go
package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

func main() {
	e := echo.New()
	e.Use(echo.WrapMiddleware(func(c echo.Context) error {
		// Extract the credentials from HTTP request header and perform a security
		// check

		// For invalid credentials
		return echo.NewHTTPError(http.StatusUnauthorized)
	}))
	e.GET("/", welcome)
	e.Run(standard.New(":1323"))
}

func welcome(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome!")
}
```

See how [HTTPErrorHandler](/guide/customization#http-error-handler) handles it.
