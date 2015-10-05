---
title: Customization
menu:
  main:
    parent: guide
    weight: 2
---

### HTTP error handler

`Echo.SetHTTPErrorHandler(h HTTPErrorHandler)`

Registers a custom `Echo.HTTPErrorHandler`.

Default handler rules:

- If error is of type `Echo.HTTPError` it sends HTTP response with status code `HTTPError.Code`
and message `HTTPError.Message`.
- Else it sends `500 - Internal Server Error`.
- If debug mode is enabled, it uses `error.Error()` as status message.

### Debug

`Echo.SetDebug(on bool)`

Enables/disables debug mode.

### Disable colored log

`Echo.DisableColoredLog()`

### StripTrailingSlash

StripTrailingSlash enables removing trailing slash from the request path.

`e.StripTrailingSlash()`
