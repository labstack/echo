+++
title = "Error Handling"
description = "Error handling in Echo"
[menu.main]
  name = "Error Handling"
  parent = "guide"
+++

Echo advocates for centralized HTTP error handling by returning error from middleware
and handlers. Centralized error handler allows us to log errors to external services
from a unified location and send a customized HTTP response to the client.

You can return a standard `error` or `echo.*HTTPError`.

For example, when basic auth middleware finds invalid credentials it returns
401 - Unauthorized error, aborting the current HTTP request.

```go
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
  return func(c echo.Context) error {
    // Extract the credentials from HTTP request header and perform a security
    // check

    // For invalid credentials
    return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")

    // For valid credentials call next
    // return next(c)
  }
})
```

You can also use `echo.NewHTTPError()` without a message, in that case status text is used
as an error message. For example, "Unauthorized".

## Default HTTP Error Handler

Echo provides a default HTTP error handler which sends error in a JSON format.

```json
{
  "message": "error connecting to redis"
}
```

For a standard `error`, response is sent as `500 - Internal Server Error`; however,
if you are running in a debug mode, the original error message is sent. If error
is `*HTTPError`, response is sent with the provided status code and message.
If logging is on, the error message is also logged.

## Custom HTTP Error Handler

For most cases default error handler should be sufficient; however, a custom HTTP
error handler can come handy if you want to capture different type of errors and
take action accordingly e.g. send notification email or log error to a centralized
system. You can also send customized responses to the clients e.g. error page or
just a JSON response.
