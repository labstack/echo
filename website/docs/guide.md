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

Echo follows [Semantic Versioning](http://semver.org) managed through GitHub releases.
Specific version of Echo can be installed using any [package manager](https://github.com/avelino/awesome-go#package-management).

## Customization

### Max path parameters

`echo.MaxParam(n uint8)`

Sets the maximum number of path parameters allowed for the application.
Default value is **5**, [good enough](https://github.com/interagent/http-api-design#minimize-path-nesting)
for many use cases. Restricting path parameters allows us to use memory efficiently.

### Not found handler

`echo.NotFoundHandler(h Handler)`

Registers a custom NotFound handler used by
router in case it doesn't find any registered handler for HTTP method and path.

Default handler sends 404 "Not Found" response.

### HTTP error handler

`echo.HTTPErrorHandler(h HTTPErrorHandler)`

Registers a centralized HTTP error handler.

Default http error handler sends 500 "Internal Server Error" response.

## Routing

Echo's router is [fast, optimized](https://github.com/labstack/echo#benchmark) and
flexible. It's based on [redix tree](http://en.wikipedia.org/wiki/Radix_tree)
data structure which makes routing lookup really fast. It leverages
[sync pool](https://golang.org/pkg/sync/#Pool) to reuse memory and achieve
zero dynamic memory allocation with no garbage collection.

Routes can be registered for any HTTP method, path and handler. For example, code
below registers a route for method `GET`, path `/hello` and a handler which sends
`Hello!` response.

```go
echo.Get("/hello", func(*echo.Context) {
	c.String(http.StatusOK, "Hello!")
})
```

<!-- ### Groups -->

### Path parameters

URL path parameters can be extracted either by name `echo.Context.Param(name string) string` or by
index `echo.Context.P(i uint8) string`. Getting parameter by index gives a slightly
better performance.

```go
echo.Get("/users/:id", func(c *echo.Context) {
	// By name
	id := c.Param("id")

	// By index
	id := c.P(0)

	c.String(http.StatusOK, id)
})
```

### Match any

Matches zero or more characters in the path. For example, pattern `/users/*` will
match

- `/users/`
- `/users/1`
- `/users/1/files/1`
- `/users/anything...`

<!-- Test it -->

### Path matching order

- Static
- Param
- Match any

<!-- Different use cases -->

### URI building

`echo.URI` can be used generate URI for any handler with specified path parameters.
It's helpful to centralize all your URI patterns which ease in refactoring your
application.

`echo.URI(h, 1)` will generate `/users/1` for the route registered below

```go
// Handler
h := func(*echo.Context) {
	c.String(http.StatusOK, "OK")
}

// Route
e.Get("/users/:id", h)
```

<!-- ## Request -->

<!-- ## Middleware -->

<!-- ## Response -->

## Static Content

### Static files

`echo.Static(path, root string)` can be used to serve static files. For example,
code below serves all files from `public/scripts` directory for any path starting
with `/scripts/`.

```go
e.Static("/scripts", "public/scripts")
```

### Serving a file

`echo.ServeFile(path, file string)` can be used to serve a file. For example, code
below serves welcome.html for path `/welcome`.

```go
e.ServeFile("/welcome", "welcome.html")
```

### Serving an index file

`echo.Index(file string)` can be used to serve index file. For example, code below
serves index.html for path `/`.

```go
e.Index("index.html")
```

<!-- ## Error Handling -->
