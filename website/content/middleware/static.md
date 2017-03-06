+++
title = "Static Middleware"
description = "Static middleware for Echo"
[menu.main]
  name = "Static"
  parent = "middleware"
+++

Static middleware can be used to serve static files from the provided root directory.

*Usage*

```go
e := echo.New()
e.Use(middleware.Static("/static"))
```

This serves static files from `static` directory. For example, a request to `/js/main.js`
will fetch and serve `static/js/main.js` file.

## Custom Configuration

*Usage*

```go
e := echo.New()
e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
  Root:   "static",
  Browse: true,
}))
```

This serves static files from `static` directory and enables directory browsing.

## Configuration

```go
StaticConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Root directory from where the static content is served.
  // Required.
  Root string `json:"root"`

  // Index file for serving a directory.
  // Optional. Default value "index.html".
  Index string `json:"index"`

  // Enable HTML5 mode by forwarding all not-found requests to root so that
  // SPA (single-page application) can handle the routing.
  // Optional. Default value false.
  HTML5 bool `json:"html5"`

  // Enable directory browsing.
  // Optional. Default value false.
  Browse bool `json:"browse"`
}
```

*Default Configuration*

```go
DefaultStaticConfig = StaticConfig{
  Skipper: DefaultSkipper,
  Index:   "index.html",
}
```
