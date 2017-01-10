+++
title = "Customization"
description = "Customizing Echo"
[menu.main]
  name = "Customization"
  parent = "guide"
  weight = 3
+++

## Debug

`Echo#Debug` can be used to enable / disable debug mode. Debug mode sets the log level
to `DEBUG`.

## Logging

### Log Output

`Echo#Logger.SetOutput(io.Writer)` can be used to set the output destination for
the logger. Default value is `os.Stdout`

To completely disable logs use `Echo#Logger.SetOutput(ioutil.Discard)` or `Echo#Logger.SetLevel(log.OFF)`

### Log Level

`Echo#Logger.SetLevel(log.Lvl)` can be used to set the log level for the logger.
Default value is `OFF`. Possible values:

- `DEBUG`
- `INFO`
- `WARN`
- `ERROR`
- `OFF`

### Custom Logger

Logging is implemented using `echo.Logger` interface which allows you to register
a custom logger using `Echo#Logger`.

## Custom Server

### Using `Echo#StartServer()`

*Example*

```go
s := &http.Server{
  Addr:         ":1323",
  ReadTimeout:  20 * time.Minute,
  WriteTimeout: 20 * time.Minute,
}
e.Logger.Fatal(e.StartServer(s))
```

### Using `http.ListenAndServe*()`


*Example*

```go
e := echo.New()
e.GET("/", func(c echo.Context) error {
  return c.JSON(http.StatusOK, "OK")
})
s := &http.Server{
  Handler: e,
  Addr:    ":1323",
}
e.Logger.Fatal(s.ListenAndServe())
```

> This setup will bypass auto-tls and graceful shutdown. 

## Disable HTTP/2

`Echo#DisableHTTP2` can be used disable HTTP/2 protocol.

## Read Timeout

`Echo#ReadTimeout` can be used to set the maximum duration before timing out read
of the request.

## Write Timeout

`Echo#WriteTimeout` can be used to set the maximum duration before timing out write
of the response.

## Shutdown Timeout

`Echo#ShutdownTimeout` can be used to set the maximum duration to wait until killing
active requests and stopping the server. If timeout is 0, the server never times
out. It waits for all active requests to finish.

## Validator

`Echo#Validator` can be used to register a validator for performing data validation
on request payload.

[Learn more](/guide/request#validate-data)

## Custom Binder

`Echo#Binder` can be used to register a custom binder for binding request payload.

[Learn more](/guide/request/#custom-binder)

## Renderer

`Echo#Renderer` can be used to register a renderer for template rendering.

[Learn more](/guide/templates)

## HTTP Error Handler

`Echo#HTTPErrorHandler` can be used to register a custom http error handler.

[Learn more](/guide/error-handling)
