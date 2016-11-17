+++
title = "Request"
description = "Handling HTTP request in Echo"
[menu.side]
  name = "Request"
  parent = "guide"
  weight = 6
+++

## Bind Request Body

To bind request body into a provided Go type use `Context#Bind(interface{})`.
The default binder supports decoding application/json, application/xml and
application/x-www-form-urlencoded payload based on Context-Type header.

*Example*

TODO

> Custom binder can be registered via `Echo#SetBinder(Binder)`

## Query Parameter

Query parameter can be retrieved by name using `Context#QueryParam(name string)`.

*Example*

```go
e.GET("/users", func(c echo.Context) error {
	name := c.QueryParam("name")
	return c.String(http.StatusOK, name)
})
```

```sh
$ curl -G -d "name=joe" http://localhost:1323/users
```

## Form Parameter

Form parameter can be retrieved by name using `Context#FormValue(name string)`.

*Example*

```go
e.POST("/users", func(c echo.Context) error {
	name := c.FormValue("name")
	return c.String(http.StatusOK, name)
})
```

```sh
$ curl -d "name=joe" http://localhost:1323/users
```

## Path Parameter

Registered path parameter can be retrieved by name `Context#Param(name string) string`.

*Example*

```go
e.GET("/users/:name", func(c echo.Context) error {
	// By name
	name := c.Param("name")

	return c.String(http.StatusOK, name)
})
```

```sh
$ curl http://localhost:1323/users/joe
```


## Handler Path

`Context#Path()` returns the registered path for the handler, it can be used in the
middleware for logging purpose.

*Example*

```go
e.Use(func(handler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		println(c.Path())
		return handler(c)
	}
})

e.GET("/users/:name", func(c echo.Context) error) {
    return c.String(http.StatusOK, name)
})
```
