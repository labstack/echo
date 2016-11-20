+++
title = "CORS Middleware"
description = "CORS middleware for Echo"
[menu.side]
  name = "CORS"
  parent = "middleware"
  weight = 5
+++

CORS middleware implements [CORS](http://www.w3.org/TR/cors) specification.
CORS gives web servers cross-domain access controls, which enable secure cross-domain
data transfers.

*Usage*

`e.Use(middleware.CORS())`

## Custom Configuration

*Usage*

```go
e := echo.New()
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
  AllowOrigins: []string{"https://labstack.com", "https://labstack.net"},
  AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
}))
```

## Configuration

```go
CORSConfig struct {
    // Skipper defines a function to skip middleware.
    Skipper Skipper

    // AllowOrigin defines a list of origins that may access the resource.
    // Optional. If request header `Origin` is set, value is []string{"<Origin>"}
    // else []string{"*"}.
    AllowOrigins []string `json:"allow_origins"`

    // AllowMethods defines a list methods allowed when accessing the resource.
    // This is used in response to a preflight request.
    // Optional. Default value DefaultCORSConfig.AllowMethods.
    AllowMethods []string `json:"allow_methods"`

    // AllowHeaders defines a list of request headers that can be used when
    // making the actual request. This in response to a preflight request.
    // Optional. Default value []string{}.
    AllowHeaders []string `json:"allow_headers"`

    // AllowCredentials indicates whether or not the response to the request
    // can be exposed when the credentials flag is true. When used as part of
    // a response to a preflight request, this indicates whether or not the
    // actual request can be made using credentials.
    // Optional. Default value false.
    AllowCredentials bool `json:"allow_credentials"`

    // ExposeHeaders defines a whitelist headers that clients are allowed to
    // access.
    // Optional. Default value []string{}.
    ExposeHeaders []string `json:"expose_headers"`

    // MaxAge indicates how long (in seconds) the results of a preflight request
    // can be cached.
    // Optional. Default value 0.
    MaxAge int `json:"max_age"`
}
```

*Default Configuration*

```go
DefaultCORSConfig = CORSConfig{
  Skipper:      defaultSkipper,
  AllowOrigins: []string{"*"},
  AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
}
```
