---
title: Routing
menu:
  side:
    parent: guide
    weight: 3
---

Echo's router is [fast, optimized]({{< relref "index.md#performance">}}) and
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

You can use `Echo.Any(path string, h Handler)` to register a handler for all HTTP methods.
To register it for some methods, use `Echo.Match(methods []string, path string, h Handler)`.


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
e.Group("/admin")
e.Use(mw.BasicAuth(func(usr, pwd string) bool {
	if usr == "joe" && pwd == "secret" {
		return true
	}
	return false
}))
```

### URI building

`Echo.URI` can be used to generate URI for any handler with specified path parameters.
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
