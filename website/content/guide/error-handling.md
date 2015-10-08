---
title: Error Handling
menu:
  side:
    parent: guide
    weight: 7
---

Echo advocates centralized HTTP error handling by returning `error` from middleware
and handlers.

It allows you to:

- Debug by writing stack trace to the HTTP response.
- Customize HTTP responses.
- Recover from panics inside middleware or handlers.

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
	e.Use(func(c *echo.Context) error {
		// Extract the credentials from HTTP request header and perform a security
		// check

		// For invalid credentials
		return echo.NewHTTPError(http.StatusUnauthorized)
	})
	e.Get("/welcome", welcome)
	e.Run(":1323")
}

func welcome(c *echo.Context) error {
	return c.String(http.StatusOK, "Welcome!")
}
```

See how [HTTPErrorHandler](/guide/customization#http-error-handler) handles it.
