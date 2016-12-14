+++
title = "Customization"
description = "Customizing Echo"
[menu.main]
  name = "Customization"
  parent = "guide"
  weight = 3
+++

## HTTP Error Handler

Default HTTP error handler sends an error as JSON with the following rules:

- If error is `Echo#HTTPError` it sends HTTP response with status code `HTTPError.Code`
and message `HTTPError.Message`.
- If error is `error` it sends HTTP response with status code `500 - Internal Server Error` 
and message `error.Error()`.
- It logs the error.

You can set a custom HTTP error handler using `Echo#HTTPErrorHandler`.

## Debugging

`Echo#Debug` enable/disable debug mode.

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
