# Echo Logger Middleware

## Overview

The Logger middleware for Echo is a powerful tool designed to log HTTP requests and responses. It's essential for debugging, monitoring application performance, and tracking user interactions with your Echo-based web services.

## Features

- Customizable log format
- Support for multiple output destinations
- Ability to skip logging for specific routes
- Customizable timestamp format
- Option to hide certain information (e.g., IP addresses)

## Installation

The Logger middleware is included in the Echo framework. To use it, you need to import Echo and its middleware package:

```go
import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)
```

## Basic Usage

To add the Logger middleware to your Echo instance with default settings:

```go
e := echo.New()
e.Use(middleware.Logger())
```

This will log requests and responses to the console (stdout) using the default format.

## Configuration

The Logger middleware can be customized using `LoggerConfig`. Here's an example of how to use it with custom settings:

```go
e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Format: "${time_custom} ${status} ${method} ${host}${path} ${latency_human}\n",
    CustomTimeFormat: "2006-01-02 15:04:05.00000",
}))
```

### Configuration Options

- `Skipper` func(c echo.Context) bool
  - Skipper defines a function to skip middleware execution.
- `Format` string
  - Format is the log format string.
- `CustomTimeFormat` string
  - CustomTimeFormat is the time format for ${time_custom}.
- `Output` io.Writer
  - Output is the writer where logs are written.

### Available Format Fields

- `${time_rfc3339}`
- `${time_unix}`
- `${time_unix_nano}`
- `${time_custom}`
- `${id}`
- `${remote_ip}`
- `${host}`
- `${method}`
- `${uri}`
- `${path}`
- `${protocol}`
- `${referer}`
- `${user_agent}`
- `${status}`
- `${error}`
- `${latency}`
- `${latency_human}`
- `${bytes_in}`
- `${bytes_out}`
- `${header:<NAME>}`
- `${query:<NAME>}`
- `${form:<NAME>}`

## Advanced Usage

### Custom Output

To log to a file instead of the console:

```go
logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    e.Logger.Fatal(err)
}
e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Output: logFile,
}))
```

### Skipping Specific Routes

To skip logging for certain routes:

```go
e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Skipper: func(c echo.Context) bool {
        return c.Path() == "/health"
    },
}))
```

### Custom Log Format

For a more detailed log format:

```go
e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Format: "${time_rfc3339_nano} ${id} ${remote_ip} ${host} ${method} ${uri} ${user_agent}" +
        " ${status} ${error} ${latency_human}" +
        " ${bytes_in} ${bytes_out}\n",
}))
```

## Best Practices

1. Use a custom format that includes all necessary information for your use case.
2. Consider using a separate log file for HTTP requests to avoid cluttering your main application logs.
3. In production environments, use a log aggregation service to centralize and analyze your logs.
4. Be mindful of sensitive information in your logs, especially when logging headers or form data.

## Performance Considerations

While the Logger middleware is designed to be efficient, logging every request can impact performance in high-traffic scenarios. Consider the following:

- Use sampling in production for very high-traffic services.
- Benchmark your application with and without the Logger middleware to understand its impact.

## Conclusion

The Echo Logger middleware is a flexible and powerful tool for monitoring and debugging your web applications. By customizing its configuration, you can tailor the logging to your specific needs while maintaining high performance.