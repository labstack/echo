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
)

func main() {
	e := echo.New()
	e.Use(func(handler echo.HandlerFunc) echo.HandlerFunc {
		// Extract the credentials from HTTP request header and perform a security
		// check
		
		// For invalid credentials
		return func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
	})

	e.GET("/welcome", welcome)
	if err := e.Start(":1323"); err != nil {
		e.Logger.Fatal(err.Error())
	}
}

func welcome(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome!")
}
```

See how [HTTPErrorHandler](/guide/customization#http-error-handler) handles it.
