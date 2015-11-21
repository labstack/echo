---
title: Request
menu:
  side:
    parent: guide
    weight: 5
---

### Handler path

`Context.Path()` returns the registered path for a handler, it can be used in the middleware for logging purpose.

*Example*

```go
e.Use(func(c *echo.Context) error {
    println(c.Path()) // Prints `/users/:name`
    return nil
})
e.Get("/users/:name", func(c *echo.Context) error) {
    return c.String(http.StatusOK, name)
})
```

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
