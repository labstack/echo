+++
title = "Middleware"
description = "Echo middleware"
type = "middleware"
[menu.main]
  name = "Middleware"
  pre = "<i class='fa fa-filter'></i>"
  weight = 2
  identifier = "middleware"
  url = "/middleware"
+++

## Overview

Middleware is a function chained in the HTTP request-response cycle with access
to `Echo#Context` which it uses to perform a specific action, for example, logging
every request or limiting the number of requests.

Handler is processed in the end after all middleware are finished executing.

## Levels

### Root Level (Before router)

`Echo#Pre()` can be used to register a middleware which is executed before router
processes the request. It is helpful to make any changes to the request properties,
for example, adding or removing a trailing slash from the path so it matches the
route.

The following built-in middleware should be registered at this level:

- HTTPSRedirect
- HTTPSWWWRedirect
- WWWRedirect
- NonWWWRedirect
- AddTrailingSlash
- RemoveTrailingSlash
- MethodOverride

> As router has not processed the request, middleware at this level won't
have access to any path related API from `echo.Context`.

### Root Level (After router)

Most of the time you will register a middleware at this level using `Echo#Use()`.
This middleware is executed after router processes the request and has full access
to `echo.Context` API.

The following built-in middleware should be registered at this level:

- BodyLimit
- Logger
- Gzip
- Recover
- BasicAuth
- JWTAuth
- Secure
- CORS
- Static

### Group Level

When creating a new group, you can register middleware just for that group. For
example, you can have an admin group which is secured by registering a BasicAuth
middleware for it.

*Usage*

```go
e := echo.New()
admin := e.Group("/admin", middleware.BasicAuth())
```

You can also add a middleware after creating a group via `admin.Use()`.

### Route Level

When defining a new route, you can optionally register middleware just for it.

*Usage*

```go
e := echo.New()
e.GET("/", <Handler>, <Middleware...>)
```

## Skipping Middleware

There are cases when you would like to skip a middleware based on some condition,
for that each middleware has an option to define a function `Skipper func(c echo.Context) bool`.

*Usage*

```go
e := echo.New()
e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
	Skipper: func(c echo.Context) bool {
		if strings.HasPrefix(c.Request().Host(), "localhost") {
			return true
		}
		return false
	},
}))
```

Example above skips Logger middleware when request host starts with localhost.

## [Writing Custom Middleware]({{< ref "cookbook/middleware.md">}})
