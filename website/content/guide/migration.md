+++
title = "Migration"
description = "Migration"
[menu.side]
  name = "Migration"
  parent = "guide"
  weight = 2
+++

## V3

### Change Log

- Automatic TLS certificates via [Let's Encrypt](https://letsencrypt.org/)
- Built-in support for graceful shutdown
- Dropped static middleware in favor of `Echo#Static`
- Utility functions to wrap standard handler and middleware
- `Map` type as shorthand for `map[string]interface{}`
- Context now wraps standard net/http Request and Response
- New configuration
	- `Echo#ShutdownTimeout`
	- `Echo#DisableHTTP2`
- New API
	- `Echo#Start()`
	- `Echo#StartTLS()`
	- `Echo#StartAutoTLS()`
	- `Echo#StartServer()`
    - `Context#Scheme()`
    - `Context#RealIP()`
    - `Context#IsTLS()`
- Exposed the following properties instead of setter / getter functions on `Echo` instance:
	- `Binder`
	- `Renderer`
	- `HTTPErrorHandler`
	- `Debug`
	- `Logger`
- Dropped API
	- `Echo#Run()`
	- `Context#P()`
- Dropped standard `Context` support 
- Dropped support for `fasthttp`
- Dropped deprecated API
- Updated docs and fixed numerous issues

### [Recipes](/recipes/hello-world)
