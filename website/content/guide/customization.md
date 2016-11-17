+++
title = "Customization"
description = "Customizing Echo"
[menu.side]
  name = "Customization"
  parent = "guide"
  weight = 3
+++

## HTTP Error Handler

Default HTTP error handler rules:

- If error is of type `Echo#HTTPError` it sends HTTP response with status code `HTTPError.Code`
and message `HTTPError.Message`.
- Else it sends `500 - Internal Server Error`.
- If debug mode is enabled, it uses `error.Error()` as status message.

You can also set a custom HTTP error handler using `Echo#HTTPErrorHandler`.

## Debugging

`Echo#Debug` enables/disables debug mode.

## Logging

### Log Output

`Echo#Logger.SetOutput(io.Writer)` sets the output destination for the logger.
Default value `os.Stdout`

To completely disable logs use `Echo#Logger.SetOutput(io.Discard)` or `Echo#Logger.SetLevel(log.OFF)`

### Log Level

`Echo#Logger.SetLevel(log.Lvl)`

SetLogLevel sets the log level for the logger. Default value `OFF`.
Possible values:

- `DEBUG`
- `INFO`
- `WARN`
- `ERROR`
- `OFF`

You can also set a custom logger using `Echo#Logger`.
