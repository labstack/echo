+++
title = "Logger Middleware"
description = "Logger middleware for Echo"
[menu.main]
  name = "Logger"
  parent = "middleware"
  weight = 5
+++

Logger middleware logs the information about each HTTP request.

*Usage*

`e.Use(middleware.Logger())`

*Sample Output*

```js
{"time":"2017-01-11T19:58:51.322299983-08:00","remote_ip":"::1","uri":"/","host":"localhost:1323","method":"GET","status":200,"latency":10667,"latency_human":"10.667µs","bytes_in":0,"bytes_out":2}
```

## Custom Configuration

*Usage*

```go
e := echo.New()
e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
  Fields: []string{"method", "uri", "status"},
}))
```

Example above uses a `Format` which logs request method and request URI.

*Sample Output*

```sh
{"uri":"/","method":"GET","status":200,"bytes_in":0,"bytes_out":0}
```

## Configuration

```go
LoggerConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Availabe logger fields:
  //
  // - time
  // - id (Request ID - Not implemented)
  // - remote_ip
  // - uri
  // - host
  // - method
  // - path
  // - referer
  // - user_agent
  // - status
  // - latency (In nanosecond)
  // - latency_human (Human readable)
  // - bytes_in (Bytes received)
  // - bytes_out (Bytes sent)
  // - header:<name>
  // - query:<name>
  // - form:<name>

  // Optional. Default value DefaultLoggerConfig.Fields.
  Fields []string `json:"fields"`

  // Output is where logs are written.
  // Optional. Default value &Stream{os.Stdout}.
  Output db.Logger
}
```

*Default Configuration*

```go
DefaultLoggerConfig = LoggerConfig{
  Skipper: defaultSkipper,
  Fields: []string{
    "time",
    "remote_ip",
    "host",
    "method",
    "uri",
    "status",
    "latency",
    "latency_human",
    "bytes_in",
    "bytes_out",
  },
  Output: &Stream{os.Stdout},
}
```
