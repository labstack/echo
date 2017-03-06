+++
title = "Migration"
description = "Migration"
[menu.main]
  name = "Migration"
  parent = "guide"
+++

## Change Log

- Automatic TLS certificates via [Let's Encrypt](https://letsencrypt.org/)
- Built-in support for graceful shutdown
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
    - `Echo#Shutdown()`
    - `Echo#ShutdownTLS()`
    - `Context#Scheme()`
    - `Context#RealIP()`
    - `Context#IsTLS()`
- Exposed the following properties instead of setter / getter functions on `Echo` instance:
	- `Binder`
	- `Renderer`
	- `HTTPErrorHandler`
	- `Debug`
	- `Logger`
- Enhanced redirect and CORS middleware
- Dropped static middleware in favor of `Echo#Static`
- Dropped API
	- `Echo#Run()`
	- `Context#P()`
- Dropped standard `Context` support 
- Dropped support for `fasthttp`
- Dropped deprecated API
- Moved `Logger` interface to root level
- Moved website and examples to the main repo
- Updated docs and fixed numerous issues

## [Cookbook](/cookbook)
