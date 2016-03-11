---
title: Migrating
menu:
  side:
    parent: guide
    weight: 2
---

### Migrating from v1

#### Change log

- Good news, 85% of the API remains the same.
- `Engine` interface to abstract `HTTP` server implementation, allowing
us to use HTTP servers beyond Go standard library. It currently supports standard HTTP server and [FastHTTP](https://github.com/valyala/fasthttp).
- Context, Request and Response are converted to interfaces. [More...](https://github.com/labstack/echo/issues/146)
- Handler signature is changed to `func (c echo.Context) error`.
- Moved API's for serving static files into middleware.
    - `Echo#Index`
    - `Echo#Favicon`
    - `Echo#Static`
    - `Echo#ServeDir`
    - `Echo#ServeFile`
- Dropped auto wrapping of handler and middleware to enforce compile time check.
- Handler only accepts `Echo#Handler` interface.
- Middleware only accepts `Echo#Middleware` interface.
- `Echo#HandlerFunc` adapter to use ordinary functions as handlers.
- `Echo#MiddlewareFunc` adapter to use ordinary functions as middleware.
- Middleware is run before hitting the router, which doesn't require `Echo#Hook` API as
it can be achieved via middleware.
- Ability to define middleware at route level.
- `Echo#HTTPError` exposed it's fields `Code` and `Message`.

#### How?

Quite easy, browse through [recipes](/recipes/hello-world) freshly converted to v2.
