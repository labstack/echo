+++
title = "JWT Middleware"
description = "JWT middleware for Echo"
[menu.main]
  name = "JWT"
  parent = "middleware"
+++

JWT provides a JSON Web Token (JWT) authentication middleware.

- For valid token, it sets the user in context and calls next handler.
- For invalid token, it sends "401 - Unauthorized" response.
- For missing or invalid `Authorization` header, it sends "400 - Bad Request".

*Usage*

`e.Use(middleware.JWT([]byte("secret"))`

## Custom Configuration

*Usage*

```go
e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
  SigningKey: []byte("secret"),
  TokenLookup: "query:token",
}))
```

## Configuration

```go
// JWTConfig defines the config for JWT middleware.
JWTConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Signing key to validate token.
  // Required.
  SigningKey interface{}

  // Signing method, used to check token signing method.
  // Optional. Default value HS256.
  SigningMethod string

  // Context key to store user information from the token into context.
  // Optional. Default value "user".
  ContextKey string

  // Claims are extendable claims data defining token content.
  // Optional. Default value jwt.MapClaims
  Claims jwt.Claims

  // TokenLookup is a string in the form of "<source>:<name>" that is used
  // to extract token from the request.
  // Optional. Default value "header:Authorization".
  // Possible values:
  // - "header:<name>"
  // - "query:<name>"
  // - "cookie:<name>"
  TokenLookup string

  // AuthScheme to be used in the Authorization header.
  // Optional. Default value "Bearer".
  AuthScheme string
}
```

*Default Configuration*

```go
DefaultJWTConfig = JWTConfig{
  Skipper:       DefaultSkipper,
  SigningMethod: AlgorithmHS256,
  ContextKey:    "user",
  TokenLookup:   "header:" + echo.HeaderAuthorization,
  AuthScheme:    "Bearer",
  Claims:        jwt.MapClaims{},
}
```

## [Example]({{< ref "cookbook/jwt.md">}})
