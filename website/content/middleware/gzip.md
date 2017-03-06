+++
title = "Gzip Middleware"
description = "Gzip middleware for Echo"
[menu.main]
  name = "Gzip"
  parent = "middleware"
+++

Gzip middleware compresses HTTP response using gzip compression scheme.

*Usage*

`e.Use(middleware.Gzip())`

## Custom Configuration

*Usage*

```go
e := echo.New()
e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
  Level: 5,
}))
```

## Configuration

```go
GzipConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Gzip compression level.
  // Optional. Default value -1.
  Level int `json:"level"`
}
```

*Default Configuration*

```go
DefaultGzipConfig = GzipConfig{
  Skipper: DefaultSkipper,
  Level:   -1,
}
```
