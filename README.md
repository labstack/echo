# *NOTICE*

- Master branch, website and godoc now points to Echo v2.
- It is advisable to migrate to v2. (https://labstack.com/echo/guide/migrating)
- Looking for v1?
	- Installation: Use a package manager (https://github.com/Masterminds/glide, it's nice!) to get stable v1 release/commit or use `go get gopkg.in/labstack/echo.v1`.
	- Godoc: https://godoc.org/gopkg.in/labstack/echo.v1
	- Docs: https://github.com/labstack/echo/tree/v1.4/website/content

# [Echo](http://labstack.com/echo) [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/labstack/echo) [![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/labstack/echo/master/LICENSE) [![Build Status](http://img.shields.io/travis/labstack/echo.svg?style=flat-square)](https://travis-ci.org/labstack/echo) [![Coverage Status](http://img.shields.io/coveralls/labstack/echo.svg?style=flat-square)](https://coveralls.io/r/labstack/echo) [![Join the chat at https://gitter.im/labstack/echo](https://img.shields.io/badge/gitter-join%20chat-brightgreen.svg?style=flat-square)](https://gitter.im/labstack/echo)

**A fast and unfancy micro web framework for Go.**

## Features

- Fast HTTP router which smartly prioritize routes.
- Run with standard HTTP server or FastHTTP server.
- Extensible middleware framework.
- Router groups with nesting.
- Handy functions to send variety of HTTP responses.
- Centralized HTTP error handling.
- Template rendering with any template engine.

## Performance

Based on [vishr/go-http-routing-benchmark] (https://github.com/vishr/go-http-routing-benchmark), June 5, 2015.

![Performance](http://i.imgur.com/hB2qdRS.png)

## Getting Started

### Installation

```sh
$ go get github.com/labstack/echo/...
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

### Next

- Browse [recipes](https://labstack.com/echo/recipes/hello-world)
- Head over to [guide](https://labstack.com/echo/guide/installation)

### Need help?

- [Hop on to chat](https://gitter.im/labstack/echo)
- [Open an issue](https://github.com/labstack/echo/issues/new)

## Want to support us?

- :star: the project
- [Donate](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=JD5R56K84A8G4&lc=US&item_name=LabStack&item_number=echo&currency_code=USD&bn=PP-DonationsBF:btn_donate_LG.gif:NonHosted)
- :earth_americas: spread the word 
- [Contribute](#contribute) to the project

## Contribute

**Use issues for everything**

- Report issues
- Discuss on chat before sending a pull request
- Suggest new features or enhancements
- Improve/fix documentation

## Credits
- [Vishal Rana](https://github.com/vishr) - Author
- [Nitin Rana](https://github.com/nr17) - Consultant
- [Contributors](https://github.com/labstack/echo/graphs/contributors)

## License

[MIT](https://github.com/labstack/echo/blob/master/LICENSE)
