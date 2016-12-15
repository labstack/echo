+++
title = "Customization"
description = "Customizing Echo"
[menu.main]
  name = "Customization"
  parent = "guide"
  weight = 3
+++

## HTTP Error Handler

`Echo#HTTPErrorHandler` can be used to set custom http error handler.

[Learn more](/guide/error-handling)

## Debugging

`Echo#Debug` can be used to enable / disable debug mode.

## Logging

### Log Output

`Echo#Logger.SetOutput(io.Writer)` can be used to set the output destination for
the logger. Default value is `os.Stdout`

To completely disable logs use `Echo#Logger.SetOutput(io.Discard)` or `Echo#Logger.SetLevel(log.OFF)`

### Log Level

`Echo#Logger.SetLevel(log.Lvl)` can be used to set the log level for the logger.
Default value `OFF`. Possible values:

- `DEBUG`
- `INFO`
- `WARN`
- `ERROR`
- `OFF`

Logging is implemented using `echo.Logger` interface which allows you to use a
custom logger. Custom logger can be set using `Echo#Logger`.

