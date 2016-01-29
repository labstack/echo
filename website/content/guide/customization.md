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

Enable/disable debug mode.

### Log prefix

`echo#SetLogPrefix(prefix string)`

SetLogPrefix sets the prefix for the logger. Default value is `echo`.

### Log output

`echo#SetLogOutput(w io.Writer)`

SetLogOutput sets the output destination for the logger. Default value is `os.Stdout`

### Log level

`echo#SetLogLevel(l log.Level)`

SetLogLevel sets the log level for the logger. Default value is `log.INFO`.

### HTTP2

`echo#HTTP2(on bool)`

Enable/disable HTTP2 support.

### Auto index

`Echo#AutoIndex(on bool)`

Enable/disable automatically creating an index page for the directory.

*Example*

```go
e := echo.New()
e.AutoIndex(true)
e.ServeDir("/", "/Users/vr/Projects/echo")
e.Run(":1323")
```

Browse to `http://localhost:1323/` to see the directory listing.

### Hook

`Echo#Hook(h http.HandlerFunc)`

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
