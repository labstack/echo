+++
title = "Migrating"
description = "Migrating from Echo v1 to v2"
[menu.side]
  name = "Migrating"
  parent = "guide"
  weight = 2
+++

## Migrating from v1

### Change Log

- Good news, 85% of the API remains the same.
- `Engine` interface to abstract `HTTP` server implementation, allowing
us to use HTTP servers beyond Go standard library. It currently supports standard and [fasthttp](https://github.com/valyala/fasthttp) server.
- Context, Request and Response are converted to interfaces. [More...](https://github.com/labstack/echo/issues/146)
- Handler signature is changed to `func (c echo.Context) error`.
- Dropped auto wrapping of handler and middleware to enforce compile time check.
- APIs to run middleware before or after the router, which doesn't require `Echo#Hook` API now.
- Ability to define middleware at route level.
- `Echo#HTTPError` exposed it's fields `Code` and `Message`.
- Option to specify log format in logger middleware and default logger.

#### API

v1 | v2
--- | ---
`Context#Query()` | `Context#QueryParam()`
`Context#Form()`  | `Context#FormValue()`

### FAQ

Q. How to access original objects from interfaces?

A. Only if you need to...

```go
// `*http.Request`
c.Request().(*standard.Request).Request

// `*http.URL`
c.Request().URL().(*standard.URL).URL

// Request `http.Header`
c.Request().Header().(*standard.Header).Header

// `http.ResponseWriter`
c.Response().(*standard.Response).ResponseWriter

// Response `http.Header`
c.Response().Header().(*standard.Header).Header
```

Q. How to use standard handler and middleware?

A.

```go
package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

// Standard middleware
func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("standard middleware")
		next.ServeHTTP(w, r)
	})
}

// Standard handler
func handler(w http.ResponseWriter, r *http.Request) {
	println("standard handler")
}

func main() {
	e := echo.New()
	e.Use(standard.WrapMiddleware(middleware))
	e.GET("/", standard.WrapHandler(http.HandlerFunc(handler)))
	e.Run(standard.New(":1323"))
}
```

### Next?

- Browse through [recipes](/recipes/hello-world) freshly converted to v2.
- Read documentation and dig into test cases.
