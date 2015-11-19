---
title: Customization
menu:
  side:
    parent: guide
    weight: 2
---

### HTTP error handler

`Echo#SetHTTPErrorHandler(h HTTPErrorHandler)`

Registers a custom `Echo#HTTPErrorHandler`.

Default handler rules:

- If error is of type `Echo#HTTPError` it sends HTTP response with status code `HTTPError.Code`
and message `HTTPError.Message`.
- Else it sends `500 - Internal Server Error`.
- If debug mode is enabled, it uses `error.Error()` as status message.

### Debug

`Echo#SetDebug(on bool)`

Enables/disables debug mode.

### Disable colored log

`Echo#DisableColoredLog()`

### Hook

`Echo#Hook(http.HandlerFunc)`

Hook registers a callback which is invoked from `Echo#ServerHTTP` as the first
statement. Hook is useful if you want to modify response/response objects even
before it hits the router or any middleware.

For example, the following hook strips the trailing slash from the request path.

```go
e.Hook(func(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    l := len(path) - 1
    if path != "/" && path[l] == '/' {
        r.URL.Path = path[:l]
    }
})
```
