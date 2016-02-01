---
title: Index
---

# ![Echo](/images/echo.svg) Echo

A fast and unfancy micro web framework for Go.

---

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
- Handy functions to send variety of HTTP response:
    - HTML
    - HTML via templates
    - String
    - JSON
    - JSONP
    - XML
    - File
    - NoContent
    - Redirect
    - Error
- Build-in support for:
	- Favicon
	- Index file
	- Static files
	- WebSocket
- Centralized HTTP error handling.
- Customizable HTTP request binding function.
- Customizable HTTP response rendering function, allowing you to use any HTML template engine.

## Performance

<iframe width="600" height="371" seamless frameborder="0" scrolling="no" src="https://docs.google.com/spreadsheets/d/1phsG_NPmEOaTVTw6lasK3CeEwBlbkhzAWPiyrBznm1g/pubchart?oid=178095723&amp;format=interactive"></iframe>

## Echo System

### Who's using Echo?

- [LabStack](https://labstack.com)
- [ShowChampions](https://showchampions.photoserve.co)
- [deferpanic](https://deferpanic.com)
- [Center for Open Science](https://cos.io)
- [SeeSaw Labs](http://www.seesawlabs.com)
- [Kyäni](http://www.kyani.net)
- [Carrot Creative](http://carrot.is)
- [EurekaMetrics](http://eurekametrics.com)
- [Coursella](https://www.coursella.com)
- [blue Vanilla](https://www.bleuvanille.fr)
- [ImPlaces](http://www.implaces.com)
- [Gomoku](http://gomoku.thoughtsfromplac.es)
- [DrinkIn](https://drinkin.com)
- [PodBaby](https://podbaby.me)

### Community created packages around Echo

- [echo-logrus](https://github.com/deoxxa/echo-logrus)
- [go_middleware](https://github.com/rightscale/go_middleware)
- [permissions2](https://github.com/xyproto/permissions2)
- [permissionbolt](https://github.com/xyproto/permissionbolt)
- [echo-middleware](https://github.com/syntaqx/echo-middleware)
- [dpecho](https://github.com/deferpanic/dpecho)
- [echosentry](https://github.com/01walid/echosentry)
- [go-starter-kit](https://github.com/olebedev/go-starter-kit)

[Want to get listed?](https://github.com/labstack/echo/issues/295)

## Getting Started

### Installation

```sh
$ go get github.com/labstack/echo
```

### Hello, World!

Create `server.go`

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

Start server

```sh
$ go run server.go
```

Browse to [http://localhost:1323](http://localhost:1323) and you should see
Hello, World! on the page.

### Next?
- Browse [recipes](https://github.com/labstack/echo/tree/master/recipes)
- Head over to [guide]({{< relref "guide/installation.md" >}})

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
