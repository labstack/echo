---
title: Migrating
menu:
  side:
    parent: guide
    weight: 2
---

### Migrating from v1

#### What got changed?

- Echo now uses `Engine` interface to abstract `HTTP` server implementation, allowing
us to use HTTP servers beyond Go standard library, supports standard HTTP server and [FastHTTP](https://github.com/valyala/fasthttp).
- Context, Request and Response as an interface, enabling adding your own functions and easy testing. [More...](https://github.com/labstack/echo/issues/146)
- Moved API's for serving static files into middleware.
    - `Echo#Index`
    - `Echo#Favicon`
    - `Echo#Static`
    - `Echo#ServeDir`
    - `Echo#ServeFile`
- Dropped auto wrapping of handler and middleware to enforce compile time check.
- Handler only accepts `Echo#Handler` interface.
- Middleware only accepts `Echo#Middleware` interface.
- `Echo#HandlerFunc` adapter to use of ordinary functions as handlers.
- `Echo#MiddlewareFunc` adapter to use of ordinary functions as middleware.
- Middleware is run before hitting the router, which doesn't require `Echo#Hook` API as
it can be achieved via middleware.
- Ability to define middleware at route level.

#### How?

##### v1 Handler

```go
func welcome(c *echo.Context) error {
	return c.String(http.StatusOK, "Welcome!\n")
}
```

##### v2 Handler

```go
func welcome(echo.HandlerFunc(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome!\n")
})
```

##### v1 Middleware

```go
func welcome(c *echo.Context) error {
	return c.String(http.StatusOK, "Welcome!\n")
}
```

v2

```go
func welcome(echo.HandlerFunc(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome!\n")
})
```
