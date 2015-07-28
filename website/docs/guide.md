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
Specific version of Echo can be installed using a [package manager](https://github.com/avelino/awesome-go#package-management).

## Customization

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

### Enable/Disable colored log 

`Echo.ColoredLog(on bool)` 

## Routing

Echo's router is [fast, optimized](https://github.com/labstack/echo#benchmark) and
flexible. It's based on [radix tree](http://en.wikipedia.org/wiki/Radix_tree) data
structure which makes route lookup really fast. Router leverages [sync pool](https://golang.org/pkg/sync/#Pool)
to reuse memory and achieve zero dynamic memory allocation with no GC overhead.

Routes can be registered by specifying HTTP method, path and a handler. For example,
code below registers a route for method `GET`, path `/hello` and a handler which sends
`Hello!` HTTP response.

```go
e.Get("/hello", func(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello!")
})
```

Echo's default handler is `func(*echo.Context) error` where `echo.Context` primarily
holds HTTP request and response objects. Echo also has a support for other types
of handlers.

### Match-any

Matches zero or more characters in the path. For example, pattern `/users/*` will
match:

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

Above routes would resolve in the following order:

- `/users/new`
- `/users/:id`
- `/users/1/files/*`

> Routes can be written in any order.

### Group

`Echo.Group(prefix string, m ...Middleware) *Group`

Routes with common prefix can be grouped to define a new sub-router with optional
middleware. If middleware is passed to the function, it overrides parent middleware
- helpful if you want a completely new middleware stack for the group. To add middleware
later you can use `Group.Use(m ...Middleware)`. Groups can also be nested.

In the code below, we create an admin group which requires basic HTTP authentication
for routes `/admin/*`.

```go
echo.Group("/admin")
e.Use(mw.BasicAuth(func(usr, pwd string) bool {
	if usr == "joe" && pwd == "secret" {
		return true
	}
	return false
}))
```

### URI building

`Echo.URI` can be used generate URI for any handler with specified path parameters.
It's helpful to centralize all your URI patterns which ease in refactoring your
application.

`e.URI(h, 1)` will generate `/users/1` for the route registered below

```go
// Handler
h := func(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

// Route
e.Get("/users/:id", h)
```

## Middleware

Middleware is a function which is chained in the HTTP request-response cycle. Middleware
has access to the request and response objects which it utilizes to perform a specific
action, for example, logging every request.

### Logger

Logs each HTTP request with method, path, status, response time and bytes served.

*Example*

```go
e.Use(Logger())

// Output: `2015/06/07 18:16:16 GET / 200 13.238µs 14`
```

### BasicAuth

BasicAuth middleware provides an HTTP basic authentication.

- For valid credentials it calls the next handler in the chain.
- For invalid Authorization header it sends "404 - Bad Request" response.
- For invalid credentials, it sends "401 - Unauthorized" response.

*Example*

```go
e.Group("/admin")
e.Use(mw.BasicAuth(func(usr, pwd string) bool {
	if usr == "joe" && pwd == "secret" {
		return true
	}
	return false
}))
```

### Gzip

Gzip middleware compresses HTTP response using gzip compression scheme.

*Example*

```go
e.Use(mw.Gzip())
```

### Recover

Recover middleware recovers from panics anywhere in the chain and handles the control
to the centralized [HTTPErrorHandler](#error-handling).

*Example*

```go
e.Use(mw.Recover())
```

### StripTrailingSlash

StripTrailingSlash middleware removes the trailing slash from request path.

*Example*

```go
e.Use(mw.StripTrailingSlash())
```

[Examples](https://github.com/labstack/echo/tree/master/examples/middleware)

## Request

### Path parameter

Path parameter can be retrieved either by name `Context.Param(name string) string`
or by index `Context.P(i int) string`. Getting parameter by index gives a slightly
better performance.

*Example*

```go
e.Get("/users/:name", func(c *echo.Context) error {
	// By name
	name := c.Param("name")

	// By index
	name := c.P(0)

	return c.String(http.StatusOK, name)
})
```

```sh
$ curl http://localhost:1323/users/joe
```

### Query parameter

Query parameter can be retrieved by name using `Context.Query(name string)`.

*Example*

```go
e.Get("/users", func(c *echo.Context) error {
	name := c.Query("name")
	return c.String(http.StatusOK, name)
})
```

```sh
$ curl -G -d "name=joe" http://localhost:1323/users
```

### Form parameter

Form parameter can be retrieved by name using `Context.Form(name string)`. 

*Example*

```go
e.Post("/users", func(c *echo.Context) error {
	name := c.Form("name")
	return c.String(http.StatusOK, name)
})
```

```sh
$ curl -d "name=joe" http://localhost:1323/users
```

## Response

### Template

```go
Context.Render(code int, name string, data interface{}) error
```
Renders a template with data and sends a text/html response with status code. Templates
can be registered using `Echo.SetRenderer()`, allowing us to use any templating engine.
Below is an example using Go `html/template`

Implement `echo.Render`

```go
Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
```

Pre-compile templates

```go
t := &Template{
    templates: template.Must(template.ParseGlob("public/views/*.html")),
}
```

Register templates

```go
e := echo.New()
e.SetRenderer(t)
e.Get("/hello", Hello)
```

Template `public/views/hello.html`

```html
{{define "hello"}}Hello, {{.}}!{{end}}

```

Handler

```go
func Hello(c *echo.Context) error {
	return c.Render(http.StatusOK, "hello", "World")
}
```

### JSON

```go
Context.JSON(code int, v interface{}) error
```

Sends a JSON HTTP response with status code.

### XML

```go
Context.XML(code int, v interface{}) error
```

Sends an XML HTTP response with status code.

### HTML

```go
Context.HTML(code int, html string) error
```

Sends an HTML HTTP response with status code.

### String

```go
Context.String(code int, s string) error
```

Sends a text/plain HTTP response with status code.

### Static files

`Echo.Static(path, root string)` serves static files. For example, code below serves
files from directory `public/scripts` for any request path starting with `/scripts/`.

```go
e.Static("/scripts/", "public/scripts")
```

### Serving a file

`Echo.ServeFile(path, file string)` serves a file. For example, code below serves
file `welcome.html` for request path `/welcome`.

```go
e.ServeFile("/welcome", "welcome.html")
```

### Serving an index file

`Echo.Index(file string)` serves root index page - `GET /`. For example, code below
serves root index page from file `public/index.html`.

```go
e.Index("public/index.html")
```

### Serving favicon

`Echo.Favicon(file string)` serves default favicon - `GET /favicon.ico`. For example,
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

See how [HTTPErrorHandler](#customization) handles it.

