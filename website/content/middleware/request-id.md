+++
title = "Request ID Middleware"
description = "Request ID middleware for Echo"
[menu.main]
  name = "Request ID"
  parent = "middleware"
+++

Request ID middleware generates a unique id for a request.

*Usage*

`e.Use(middleware.RequestID())`

## Custom Configuration

*Usage*

```go
e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
  Generator: func() string {
    return customGenerator()
  },
}))
```

## Configuration

```go
// RequestIDConfig defines the config for RequestID middleware.
RequestIDConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Generator defines a function to generate an ID.
  // Optional. Default value random.String(32).
  Generator func() string
}
```

*Default Configuration*

```go
DefaultRequestIDConfig = RequestIDConfig{
  Skipper:   DefaultSkipper,
  Generator: generator,
}
```
