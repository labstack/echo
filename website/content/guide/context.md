+++
title = "Context"
description = "Context in Echo"
[menu.side]
  name = "Context"
  identifier = "context"
  parent = "guide"
  weight = 5
+++

## Context

`echo.Context` represents the context of the current HTTP request. It holds request and
response reference, path, path parameters, data, registered handler and APIs to read
request and write response. Context is 100% compatible with standard `context.Context`.
As Context is an interface, it is easy to extend it with custom APIs.

#### Extending Context

**Define a custom context**

```go
type CustomContext struct {
	echo.Context
}

func (c *CustomContext) Foo() {
	println("foo")
}

func (c *CustomContext) Bar() {
	println("bar")
}
```

**Create a middleware to extend default context**

```go
e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := &CustomContext{c}
		return h(cc)
	}
})
```

> This middleware should be registered before any other middleware.

**Use in handler**

```go
e.Get("/", func(c echo.Context) error {
	cc := c.(*CustomContext)
	cc.Foo()
	cc.Bar()
	return cc.String(200, "OK")
})
```

### Standard Context

`echo.Context` embeds standard `context.Context` interface, so all it's functions
are available right from `echo.Context`.

*Example*

```go
e.GET("/users/:name", func(c echo.Context) error) {
    c.SetContext(context.WithValue(nil, "key", "val"))
    // Pass it down...
    // Use it...
    val := c.Value("key").(string)
    return c.String(http.StatusOK, name)
})
```
