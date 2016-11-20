+++
title = "BasicAuth Middleware"
description = "Basic auth middleware for Echo"
[menu.side]
  name = "BasicAuth"
  parent = "middleware"
  weight = 5
+++

BasicAuth middleware provides an HTTP basic authentication.

- For valid credentials it calls the next handler.
- For invalid credentials, it sends "401 - Unauthorized" response.
- For empty or invalid `Authorization` header, it sends "400 - Bad Request" response.

*Usage*

```go
e := echo.New()
e.Use(middleware.BasicAuth(func(username, password string) bool {
	if username == "joe" && password == "secret" {
		return true
	}
	return false
}))
```

## Custom Configuration

*Usage*

```go
e := echo.New()
e.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{},
}))
```

## Configuration

```go
BasicAuthConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Validator is a function to validate BasicAuth credentials.
  // Required.
  Validator BasicAuthValidator
}
```

*Default Configuration*

```go
DefaultBasicAuthConfig = BasicAuthConfig{
	Skipper: defaultSkipper,
}
```
