+++
title = "TrailingSlash Middleware"
description = "Trailing slash middleware for Echo"
[menu.side]
  name = "TrailingSlash"
  parent = "middleware"
  weight = 5
+++

## AddTrailingSlash Middleware

AddTrailingSlash middleware adds a trailing slash to the request URI.

*Usage*

```go
e := echo.New()
e.Pre(middleware.AddTrailingSlash())
```

## RemoveTrailingSlash Middleware

RemoveTrailingSlash middleware removes a trailing slash from the request URI.

*Usage*

```go
e := echo.New()
e.Pre(middleware.RemoveTrailingSlash())
```

## Custom Configuration

*Usage*

```go
e := echo.New()
e.Use(middleware.AddTrailingSlashWithConfig(middleware.TrailingSlashConfig{
  RedirectCode: http.StatusMovedPermanently,
}))
```

Example above will add a trailing slash to the request URI and redirect with `308 - StatusMovedPermanently`.

## Configuration

```go
TrailingSlashConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Status code to be used when redirecting the request.
  // Optional, but when provided the request is redirected using this code.
  RedirectCode int `json:"redirect_code"`
}
```

*Default Configuration*

```go
DefaultTrailingSlashConfig = TrailingSlashConfig{
  Skipper: defaultSkipper,
}
```
