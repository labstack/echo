---
title: Customization
menu:
  side:
    parent: guide
    weight: 3
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

`Echo#SetLogPrefix(prefix string)`

SetLogPrefix sets the prefix for the logger. Default value is `echo`.

### Log output

`Echo#SetLogOutput(w io.Writer)`

SetLogOutput sets the output destination for the logger. Default value is `os.Stdout`

To completely disable logs use `Echo#SetLogOutput(io.Discard)`

### Log level

`Echo#SetLogLevel(l log.Level)`

SetLogLevel sets the log level for the logger. Default value is `log.ERROR`.

### Engine

#### Standard HTTP server

```go
e.Run(standard.New(":1323"))
```

##### From TLS

```go
e.Run(standard.NewFromTLS(":1323", "<certfile>", "<keyfile>"))
```

##### From config

```go
e.Run(standard.NewFromConfig(&Config{}))
```

#### FastHTTP server

```go
e.Run(fasthttp.New(":1323"))
```

##### From TLS

```go
e.Run(fasthttp.NewFromTLS(":1323", "<certfile>", "<keyfile>"))
```


##### From config
```go
e.Run(fasthttp.NewFromConfig(&Config{}))
```

#### Configuration

##### `Address`

Address to bind.

##### `TLSCertfile`

TLS certificate file path

##### `TLSKeyfile`

TLS key file path

##### `ReadTimeout`

HTTP read timeout

##### `WriteTimeout`

HTTP write timeout
