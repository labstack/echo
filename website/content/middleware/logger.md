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
{"time":"2017-01-12T08:58:07.372015644-08:00","remote_ip":"::1","host":"localhost:1323","method":"GET","uri":"/","status":200, "latency":14743,"latency_human":"14.743µs","bytes_in":0,"bytes_out":2}
```

## Custom Configuration

*Usage*

```go
e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
  Format: "method=${method}, uri=${uri}, status=${status}\n",
}))
```

Example above uses a `Format` which logs request method and request URI.

*Sample Output*

```sh
method=GET, uri=/, status=200
```

## Configuration

```go
// LoggerConfig defines the config for Logger middleware.
LoggerConfig struct {
  // Skipper defines a function to skip middleware.
  Skipper Skipper

  // Tags to constructed the logger format.
  //
  // - time_unix
  // - time_unix_nano
  // - time_rfc3339
  // - time_rfc3339_nano
  // - id (Request ID - Not implemented)
  // - remote_ip
  // - uri
  // - host
  // - method
  // - path
  // - referer
  // - user_agent
  // - status
  // - latency (In nanoseconds)
  // - latency_human (Human readable)
  // - bytes_in (Bytes received)
  // - bytes_out (Bytes sent)
  // - header:<NAME>
  // - query:<NAME>
  // - form:<NAME>
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
  Format: `{"time":"${time_rfc3339_nano}","remote_ip":"${remote_ip}","host":"${host}",` +
    `"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
    `"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
    `"bytes_out":${bytes_out}}` + "\n",
  Output: os.Stdout
}
```
