# Echo

Build simple and performant systems!

---

## Overview

Echo is a fast HTTP router (zero dynamic memory allocation) and micro web framework in Go.

## Features

- Fast HTTP router which smartly prioritize routes.
- Extensible middleware, supports:
	- `echo.MiddlewareFunc`
	- `func(echo.HandlerFunc) echo.HandlerFunc`
	- `echo.HandlerFunc`
	- `func(*echo.Context) error`
	- `func(http.Handler) http.Handler`
	- `http.Handler`
	- `http.HandlerFunc`
	- `func(http.ResponseWriter, *http.Request)`
- Extensible handler, supports:
    - `echo.HandlerFunc`
    - `func(*echo.Context) error`
    - `http.Handler`
    - `http.HandlerFunc`
    - `func(http.ResponseWriter, *http.Request)`
- Sub-router/Groups
- Handy encoding/decoding functions.
- Build-in support for:
	- Static files
	- WebSocket
- API to serve index and favicon.
- Centralized HTTP error handling.
- Customizable request binding function.
- Customizable response rendering function, allowing you to use any HTML template engine.

## Getting Started

### Installation

```sh
$ go get github.com/labstack/echo
```

###[Hello, World!](https://github.com/labstack/echo/tree/master/examples/hello)

Create `server.go` with the following content

```go
package main

import (
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

// Handler
func hello(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())

	// Routes
	e.Get("/", hello)

	// Start server
	e.Run(":1323")
}
```

`echo.New()` returns a new instance of Echo.

`e.Use(mw.Logger())` adds logging middleware to the chain. It logs every HTTP request
made to the server, producing output

```sh
2015/06/07 18:16:16 GET / 200 13.238µs 14
```

`e.Get("/", hello)` Registers hello handler for HTTP method `GET` and path `/`, so
whenever server receives an HTTP request at `/`, hello function is called.

In hello handler `c.String(http.StatusOK, "Hello, World!\n")` sends a text/plain
HTTP response to the client with 200 status code.

`e.Run(":1323")` Starts HTTP server at network address `:1323`.

Now start the server using command

```sh
$ go run server.go
```

Browse to [http://localhost:1323](http://localhost:1323) and you should see
Hello, World! on the page.

### Next?
- Browse [examples](https://github.com/labstack/echo/tree/master/examples)
- Head over to [Guide](guide.md)

## Contribute

**Use issues for everything**

- Report issues
- Discuss before sending pull request
- Suggest new features
- Improve/fix documentation

## Credits
- [Vishal Rana](https://github.com/vishr) - Author
- [Nitin Rana](https://github.com/nr17) - Consultant
- [Contributors](https://github.com/labstack/echo/graphs/contributors)

## License

[MIT](https://github.com/labstack/echo/blob/master/LICENSE)
