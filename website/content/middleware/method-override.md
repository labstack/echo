+++
title = "Method Override Middleware"
description = "Method override middleware for Echo"
[menu.main]
  name = "Method Override"
  parent = "middleware"
  weight = 5
+++

Method override middleware checks for the overridden method from the request and
uses it instead of the original method.

For security reasons, only `POST` method can be overridden.

*Usage*

`e.Pre(middleware.MethodOverride())`

## Custom Configuration

*Usage*

```go
e := echo.New()
e.Pre(middleware.MethodOverrideWithConfig(middleware.MethodOverrideConfig{
  Getter: middleware.MethodFromForm("_method"),
}))
```

## Configuration

```go
MethodOverrideConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Getter is a function that gets overridden method from the request.
  // Optional. Default values MethodFromHeader(echo.HeaderXHTTPMethodOverride).
  Getter MethodOverrideGetter
}
```

*Default Configuration*

```go
DefaultMethodOverrideConfig = MethodOverrideConfig{
  Skipper: defaultSkipper,
  Getter:  MethodFromHeader(echo.HeaderXHTTPMethodOverride),
}
```
