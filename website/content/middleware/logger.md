+++
title = "Logger Middleware"
description = "Logger middleware for Echo"
[menu.side]
  name = "Logger"
  parent = "middleware"
  weight = 5
+++

Logger middleware logs the information about each HTTP request.

*Usage*

`e.Use(middleware.Logger())`

*Sample Output*

```js
{"time":"2016-05-10T07:02:25-07:00","remote_ip":"::1","method":"GET","uri":"/","status":200, "latency":55653,"latency_human":"55.653µs","rx_bytes":0,"tx_bytes":13}
```

## Custom Configuration

*Usage*

```go
e := echo.New()
e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
  Format: "method=${method}, uri=${uri}, status=${status}\n",
}))
```

Example above uses a `Format` which logs request method and request URI.

*Sample Output*

```sh
method=GET, uri=/hello, status=200
```

## Configuration

```go
LoggerConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Log format which can be constructed using the following tags:
  //
  // - time_rfc3339
  // - id (Request ID - Not implemented)
  // - remote_ip
  // - uri
  // - host
  // - method
  // - path
  // - referer
  // - user_agent
  // - status
  // - latency (In microseconds)
  // - latency_human (Human readable)
  // - bytes_in (Bytes received)
  // - bytes_out (Bytes sent)
  // - header:<name>
  // - query:<name>
  // - form:<name>
  //
  // Example "${remote_ip} ${status}"
  //
  // Optional. Default value DefaultLoggerConfig.Format.
  Format string `json:"format"`

  // Output is a writer where logs are written.
  // Optional. Default value os.Stdout.
  Output io.Writer
}
```

*Default Configuration*

```go
DefaultLoggerConfig = LoggerConfig{
  Skipper: defaultSkipper,
  Format: `{"time":"${time_rfc3339}","remote_ip":"${remote_ip}",` +
    `"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
    `"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
    `"bytes_out":${bytes_out}}` + "\n",
  Output: os.Stdout,
}
```
