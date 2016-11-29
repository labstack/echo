+++
title = "Context"
description = "Context in Echo"
[menu.main]
  name = "Context"
  parent = "guide"
  weight = 5
+++

`echo.Context` represents the context of the current HTTP request. It holds request and
response reference, path, path parameters, data, registered handler and APIs to read
request and write response. As Context is an interface, it is easy to extend it with
custom APIs.

## Extending Context

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
e.GET("/", func(c echo.Context) error {
	cc := c.(*CustomContext)
	cc.Foo()
	cc.Bar()
	return cc.String(200, "OK")
})
```
