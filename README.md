# *NOTICE*

- Master branch, website and godoc now points to Echo v2.
- It is advisable to migrate to v2. (https://labstack.com/echo/guide/migrating/)
- Looking for v1?
	- Installation: Use a package manager (https://github.com/Masterminds/glide, it's nice!) to get stable v1 release/commit or use `go get gopkg.in/labstack/echo.v1`.
	- Godoc: https://godoc.org/gopkg.in/labstack/echo.v1
	- Docs: https://github.com/labstack/echo/tree/v1.4/website/content

# [Echo](http://labstack.com/echo) [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/labstack/echo) [![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/labstack/echo/master/LICENSE) [![Build Status](http://img.shields.io/travis/labstack/echo.svg?style=flat-square)](https://travis-ci.org/labstack/echo) [![Coverage Status](http://img.shields.io/coveralls/labstack/echo.svg?style=flat-square)](https://coveralls.io/r/labstack/echo) [![Join the chat at https://gitter.im/labstack/echo](https://img.shields.io/badge/gitter-join%20chat-brightgreen.svg?style=flat-square)](https://gitter.im/labstack/echo)

#### Echo is a fast and unfancy web framework for Go (Golang). Up to 10x faster than the rest.

## Features

- Optimized HTTP router which smartly prioritize routes.
- Build robust and scalable RESTful APIs.
- Run with standard HTTP server or FastHTTP server.
- Group APIs.
- Extensible middleware framework.
- Define middleware at root, group or route level.
- Handy functions to send variety of HTTP responses.
- Centralized HTTP error handling.
- Template rendering with any template engine.
- Define your format for the logger.
- Highly customizable.

## Performance

- Environment:
	- Go 1.6
	- wrk 4.0.0
	- 2 GB, 2 Core (DigitalOcean)
- Test Suite: https://github.com/vishr/web-framework-benchmark
- Date: 4/4/2016

![Performance](http://i.imgur.com/fZVnK52.png)

## Quick Start

### Installation

```sh
$ go get github.com/labstack/echo/...
```

### Hello, World!

Create `server.go`

```go
package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	e.Get("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
}
```

Start server

```sh
$ go run server.go
```

Browse to [http://localhost:1323](http://localhost:1323) and you should see
Hello, World! on the page.

### Routing

```go
e.Post("/users", saveUser)
e.Get("/users/:id", getUser)
e.Put("/users/:id", updateUser)
e.Delete("/users/:id", deleteUser)
```

### Path Parameters

```go
func getUser(c echo.Context) error {
	// User ID from path `users/:id`
	id := c.Param("id")
}
```

### Query Parameters

`/show?team=x-men&member=wolverine`

```go
func show(c echo.Context) error {
	// Get team and member from the query string
	team := c.QueryParam("team")
	member := c.QueryParam("member")
}
```

### Form `application/x-www-form-urlencoded`

`POST` `/save` `name=Joe Smith, email=joe@labstack.com`

name | value
:--- | :---
name | Joe Smith
email | joe@labstack.com


```go
func save(c echo.Context) error {
	// Get name and email
	name := c.FormValue("name")
	email := c.FormParam("email")
}
```

### Form `multipart/form-data`

`POST` `/save`

name | value
:--- | :---
name | Joe Smith
email | joe@labstack.com
avatar | avatar

```go
func save(c echo.Context) error {
	// Get name and email
	name := c.FormValue("name")
	email := c.FormParam("email")

	//------------
	// Get avatar
	//------------

	avatar, err := c.FormFile("avatar")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	file, err := os.Create(file.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy
	if _, err = io.Copy(file, avatar); err != nil {
		return err
	}
}
```

### Handling Request

- Bind `JSON` or `XML` payload into Go struct based on `Content-Type` request header.
- Render response as `JSON` or `XML` with status code.

```go
type User struct {
	Name  string `json:"name" xml:"name"`
	Email string `json:"email" xml:"email"`
}

e.Post("/users", func(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, u)
	// or
	// return c.XML(http.StatusCreated, u)
})
```

### Static Content

Server any file from static directory for path `/static/*`.

```go
e.Static("/static", "static")
```

[More...](https://labstack.com/echo/guide/static-files/)

### [Template Rendering](https://labstack.com/echo/guide/templates/)

### Middleware

```go
// Root level middleware
e.Use(middleware.Logger())
e.Use(middleware.Recover())

// Group level middleware
g := e.Group("/admin")
g.Use(middleware.BasicAuth(func(username, password string) bool {
	if username == "joe" && password == "secret" {
		return true
	}
	return false
}))

// Route level middleware
track := func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		println("request to /users")
		return next(c)
	}
}
e.Get("/users", func(c echo.Context) error {
	return c.String(http.StatusOK, "/users")
}, track)
```

[More...](https://labstack.com/echo/guide/middleware/)

### Next

- Head over to [guide](https://labstack.com/echo/guide/installation/)
- Browse [recipes](https://labstack.com/echo/recipes/hello-world/)

### Need help?

- [Hop on to chat](https://gitter.im/labstack/echo)
- [Open an issue](https://github.com/labstack/echo/issues/new)

## Support Us

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
