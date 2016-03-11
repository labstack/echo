---
title: Middleware
menu:
  side:
    parent: guide
    weight: 5
---

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
g := e.Group("/admin")
e.Use(middleware.BasicAuth(func(username, password string) bool {
	if username == "joe" && password == "secret" {
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

Recover middleware recovers from panics anywhere in the chain and handles the
control to the centralized
[HTTPErrorHandler]({{< relref "guide/customization.md#http-error-handler">}}).

*Example*

```go
e.Use(middleware.Recover())
```

### [Recipes](https://github.com/labstack/echox/tree/master/recipe/middleware)
