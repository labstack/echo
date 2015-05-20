# Guide

<!---
	Some info about guide
-->

---

## Installation

Echo has been developed and tested using Go `1.4.x`

Install the latest version of Echo via `go get`

```sh
$ go get github.com/labstack/echo
```

To upgrade

```sh
$ go get -u github.com/labstack/echo
```

Echo follows [semantic versioning](http://semver.org) managed through GitHub releases.
Specific version of Echo can be installed using any [package manager](https://github.com/avelino/awesome-go#package-management).

## Customization

### Max path parameters

`Echo.SetMaxParam(n uint8)`

Sets the maximum number of path parameters allowed for the application.
Default value is **5**, [good enough](https://github.com/interagent/http-api-design#minimize-path-nesting)
for many use cases. Restricting path parameters allows us to use memory efficiently.

### HTTP error handler

`Echo.SetHTTPErrorHandler(h HTTPErrorHandler)`

Registers a custom `Echo.HTTPErrorHandler`.

Default handler rules

- If error is of type `Echo.HTTPError` it sends HTTP response with status code `HTTPError.Code`
and message `HTTPError.Message`.
- Else it sends `500 - Internal Server Error`.
- If debug mode is enabled, it uses `error.Error()` as status message.

### Debug

`Echo.SetDebug(on bool)`

Enables debug mode.

## Routing

Echo's router is [fast, optimized](https://github.com/labstack/echo#benchmark) and
flexible. It's based on [redix tree](http://en.wikipedia.org/wiki/Radix_tree)
data structure which makes routing lookup really fast. It leverages
[sync pool](https://golang.org/pkg/sync/#Pool) to reuse memory and achieve
zero dynamic memory allocation with no GC overhead.

Routes can be registered by specifying HTTP method, path and a handler. For example,
code below registers a route for method `GET`, path `/hello` and a handler which sends
`Hello!` HTTP response.

```go
echo.Get("/hello", func(*echo.Context) error {
	return c.String(http.StatusOK, "Hello!")
})
```

Echo's default handler is `func(*echo.Context) error` where `echo.Context`
primarily holds HTTP request and response objects. Echo also has a support for other
types of handlers.

<!-- TODO mention about not able to take advantage -->

### Group

*WIP*

### Path parameter

Request path parameters can be extracted either by name `echo.Context.Param(name string) string`
or by index `echo.Context.P(i uint8) string`. Getting parameter by index gives a
slightly better performance.

```go
echo.Get("/users/:id", func(c *echo.Context) error {
	// By name
	id := c.Param("id")

	// By index
	id := c.P(0)

	return c.String(http.StatusOK, id)
})
```

### Match-any

Matches zero or more characters in the path. For example, pattern `/users/*` will
match

- `/users/`
- `/users/1`
- `/users/1/files/1`
- `/users/anything...`

### Path matching order

- Static
- Param
- Match any

#### Example

```go
e.Get("/users/:id", func(c *echo.Context) error {
	return c.String(http.StatusOK, "/users/:id")
})

e.Get("/users/new", func(c *echo.Context) error {
	return c.String(http.StatusOK, "/users/new")
})

e.Get("/users/1/files/*", func(c *echo.Context) error {
	return c.String(http.StatusOK, "/users/1/files/*")
})
```

Above routes would resolve in order

- `/users/new`
- `/users/:id`
- `/users/1/files/*`

Routes can be written in any order.

<!-- Different use cases -->

### URI building

`echo.URI` can be used generate URI for any handler with specified path parameters.
It's helpful to centralize all your URI patterns which ease in refactoring your
application.

`echo.URI(h, 1)` will generate `/users/1` for the route registered below

```go
// Handler
h := func(*echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

// Route
e.Get("/users/:id", h)
```

## Middleware

[*WIP*](https://github.com/labstack/echo/tree/master/examples/middleware)

## Response

### JSON

```go
context.JSON(code int, v interface{}) error
```

Sends a JSON HTTP response with status code.

### String

```go
context.String(code int, s string) error
```

Sends a text/plain HTTP response with status code.

### HTML

```go
func (c *Context) HTML(code int, html string) error
```

Sends an HTML HTTP response with status code.

### Static files

`echo.Static(path, root string)` serves static files. For example, code below serves
files from directory `public/scripts` for any request path starting with `/scripts/`.

```go
e.Static("/scripts/", "public/scripts")
```

### Serving a file

`echo.ServeFile(path, file string)` serves a file. For example, code below serves
file `welcome.html` for request path `/welcome`.

```go
e.ServeFile("/welcome", "welcome.html")
```

### Serving an index file

`echo.Index(file string)` serves root index page - `GET /`. For example, code below
serves root index page from file `public/index.html`.

```go
e.Index("public/index.html")
```

### Serving favicon 

`echo.Favicon(file string)` serves default favicon - `GET /favicon.ico`. For example,
code below serves favicon from file `public/favicon.ico`.

```go
e.Favicon("public/favicon.ico")
```

## Error Handling

Echo advocates centralized HTTP error handling by returning `error` from middleware
and handlers.

It allows you to

- Debug by writing stack trace to the HTTP response.
- Customize HTTP responses.
- Recover from panics inside middleware or handlers.

For example, when a basic auth middleware finds invalid credentials it returns
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

See how [HTTPErrorHandler](#customization) handles it.

## Deployment

*WIP*
