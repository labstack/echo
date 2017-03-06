+++
title = "Recover Middleware"
description = "Recover middleware for Echo"
[menu.main]
  name = "Recover"
  parent = "middleware"
+++

Recover middleware recovers from panics anywhere in the chain, prints stack trace
and handles the control to the centralized
[HTTPErrorHandler]({{< ref "guide/customization.md#http-error-handler">}}).

*Usage*

`e.Use(middleware.Recover())`

## Custom Configuration

*Usage*

```go
e := echo.New()
e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
  StackSize:  1 << 10, // 1 KB
}))
```

Example above uses a `StackSize` of 1 KB and default values for `DisableStackAll`
and `DisablePrintStack`.

## Configuration

```go
RecoverConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Size of the stack to be printed.
  // Optional. Default value 4KB.
  StackSize int `json:"stack_size"`

  // DisableStackAll disables formatting stack traces of all other goroutines
  // into buffer after the trace for the current goroutine.
  // Optional. Default value false.
  DisableStackAll bool `json:"disable_stack_all"`

  // DisablePrintStack disables printing stack trace.
  // Optional. Default value as false.
  DisablePrintStack bool `json:"disable_print_stack"`
}
```

*Default Configuration*

```go
DefaultRecoverConfig = RecoverConfig{
  Skipper:           DefaultSkipper,
  StackSize:         4 << 10, // 4 KB
  DisableStackAll:   false,
  DisablePrintStack: false,
}
```
