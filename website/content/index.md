---
title: Index
---

# ![Echo](/images/echo.svg) Echo

A fast and unfancy micro web framework for Go.

---

## Features

- Fast HTTP router which smartly prioritize routes.
- Run with standard HTTP server or FastHTTP server.
- Extensible middleware framework.
- Router groups with nesting.
- Handy functions to send variety of HTTP responses.
- Centralized HTTP error handling.
- Template rendering with any template engine.

## Performance

<iframe width="600" height="371" seamless frameborder="0" scrolling="no" src="https://docs.google.com/spreadsheets/d/1phsG_NPmEOaTVTw6lasK3CeEwBlbkhzAWPiyrBznm1g/pubchart?oid=178095723&amp;format=interactive"></iframe>

## Getting Started

### Installation

```sh
$ go get github.com/labstack/echo
```

### Hello, World!

Create `main.go`

```go
package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

// Handler
func hello() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!\n")
	}
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.Get("/", hello())

	// Start server
	e.Run(standard.New(":1323"))
}
```

Start server

```sh
$ go run main.go
```

Browse to [http://localhost:1323](http://localhost:1323) and you should see
Hello, World! on the page.

### Next?
- Browse [recipes](/recipes/hello-world)
- Head over to [guide](/guide/installation")

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
