# Development Guidelines for middlewares

## Best practices:

* Do not use `panic` in middleware creator functions in case of invalid configuration.
* In case of an error in middleware function handling request avoid using `c.Error()` and returning no error instead
  because previous middlewares up in call chain could have logic for dealing with returned errors.
* Create middleware configuration structs that implement `MiddlewareConfigurator` interface so can decide if they
  want to create middleware with panics or with returning errors on configuration errors.
* When adding `echo.Context` to function type or fields make it first parameter so all functions with Context looks same.

