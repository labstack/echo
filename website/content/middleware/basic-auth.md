+++
title = "Basic Auth Middleware"
description = "Basic auth middleware for Echo"
[menu.main]
  name = "Basic Auth"
  parent = "middleware"
+++

Basic auth middleware provides an HTTP basic authentication.

- For valid credentials it calls the next handler.
- For missing or invalid credentials, it sends "401 - Unauthorized" response.

*Usage*

```go
e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (error, bool) {
	if username == "joe" && password == "secret" {
		return nil, true
	}
	return nil, false
}))
```

## Custom Configuration

*Usage*

```go
e.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{}}))
```

## Configuration

```go
BasicAuthConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Validator is a function to validate BasicAuth credentials.
  // Required.
  Validator BasicAuthValidator

  // Realm is a string to define realm attribute of BasicAuth.
  // Default value "Restricted".
  Realm string
}
```

*Default Configuration*

```go
DefaultBasicAuthConfig = BasicAuthConfig{
	Skipper: DefaultSkipper,
}
```
