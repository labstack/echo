# [Echo](http://labstack.com/echo) [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/labstack/echo) [![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/labstack/echo/master/LICENSE) [![Build Status](http://img.shields.io/travis/labstack/echo.svg?style=flat-square)](https://travis-ci.org/labstack/echo) [![Coverage Status](http://img.shields.io/coveralls/labstack/echo.svg?style=flat-square)](https://coveralls.io/r/labstack/echo) [![Join the chat at https://gitter.im/labstack/echo](https://img.shields.io/badge/gitter-join%20chat-brightgreen.svg?style=flat-square)](https://gitter.im/labstack/echo) [![Twitter](https://img.shields.io/badge/twitter-@labstack-55acee.svg?style=flat-square)](https://twitter.com/labstack)

## Don't forget to try the upcoming [v3](https://github.com/labstack/echo/tree/v3) tracked [here]( https://github.com/labstack/echo/issues/665)

#### Fast and unfancy HTTP server framework for Go (Golang). Up to 10x faster than the rest.

## Feature Overview

- Optimized HTTP router which smartly prioritize routes
- Build robust and scalable RESTful APIs
- Run with standard HTTP server or FastHTTP server
- Group APIs
- Extensible middleware framework
- Define middleware at root, group or route level
- Data binding for JSON, XML and form payload
- Handy functions to send variety of HTTP responses
- Centralized HTTP error handling
- Template rendering with any template engine
- Define your format for the logger
- Highly customizable

## Performance

- Environment:
	- Go 1.6
	- wrk 4.0.0
	- 2 GB, 2 Core (DigitalOcean)
- Test Suite: https://github.com/vishr/web-framework-benchmark
- Date: 4/4/2016

![Performance](https://i.imgur.com/fZVnK52.png)

## Quick Start

### Installation

Echo is developed and tested using Go `1.6.x` and `1.7.x`

```sh
$ go get -u github.com/labstack/echo
```

> Ideally, you should rely on a [package manager](https://github.com/avelino/awesome-go#package-management) like glide or govendor to use a specific [version](https://github.com/labstack/echo/releases) of Echo.

### Hello, World!

Create `server.go`

```go
package main

import (
	"net/http"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Run(standard.New(":1323"))
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
e.POST("/users", saveUser)
e.GET("/users/:id", getUser)
e.PUT("/users/:id", updateUser)
e.DELETE("/users/:id", deleteUser)
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

`POST` `/save`

name | value
:--- | :---
name | Joe Smith
email | joe@labstack.com


```go
func save(c echo.Context) error {
	// Get name and email
	name := c.FormValue("name")
	email := c.FormValue("email")
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
	email := c.FormValue("email")
	// Get avatar
	avatar, err := c.FormFile("avatar")
	if err != nil {
		return err
	}

	// Source
	src, err := avatar.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	dst, err := os.Create(avatar.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.HTML(http.StatusOK, "<b>Thank you!</b>")
}
```

### Handling Request

- Bind `JSON` or `XML` or `form` payload into Go struct based on `Content-Type` request header.
- Render response as `JSON` or `XML` with status code.

```go
type User struct {
	Name  string `json:"name" xml:"name" form:"name"`
	Email string `json:"email" xml:"email" form:"email"`
}

e.POST("/users", func(c echo.Context) error {
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

##### [Learn More](https://echo.labstack.com/guide/static-files)

### [Template Rendering](https://echo.labstack.com/guide/templates)

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
e.GET("/users", func(c echo.Context) error {
	return c.String(http.StatusOK, "/users")
}, track)
```

#### Built-in Middleware

Middleware | Description
:--- | :---
[BodyLimit](https://echo.labstack.com/middleware/body-limit) | Limit request body
[Logger](https://echo.labstack.com/middleware/logger) | Log HTTP requests
[Recover](https://echo.labstack.com/middleware/recover) | Recover from panics
[Gzip](https://echo.labstack.com/middleware/gzip) | Send gzip HTTP response
[BasicAuth](https://echo.labstack.com/middleware/basic-auth) | HTTP basic authentication
[JWTAuth](https://echo.labstack.com/middleware/jwt) | JWT authentication
[Secure](https://echo.labstack.com/middleware/secure) | Protection against attacks
[CORS](https://echo.labstack.com/middleware/cors) | Cross-Origin Resource Sharing
[CSRF](https://echo.labstack.com/middleware/csrf) | Cross-Site Request Forgery
[Static](https://echo.labstack.com/middleware/static) | Serve static files
[HTTPSRedirect](https://echo.labstack.com/middleware/redirect#httpsredirect-middleware) | Redirect HTTP requests to HTTPS
[HTTPSWWWRedirect](https://echo.labstack.com/middleware/redirect#httpswwwredirect-middleware) | Redirect HTTP requests to WWW HTTPS
[WWWRedirect](https://echo.labstack.com/middleware/redirect#wwwredirect-middleware) | Redirect non WWW requests to WWW
[NonWWWRedirect](https://echo.labstack.com/middleware/redirect#nonwwwredirect-middleware) | Redirect WWW requests to non WWW
[AddTrailingSlash](https://echo.labstack.com/middleware/trailing-slash#addtrailingslash-middleware) | Add trailing slash to the request URI
[RemoveTrailingSlash](https://echo.labstack.com/middleware/trailing-slash#removetrailingslash-middleware) | Remove trailing slash from the request URI
[MethodOverride](https://echo.labstack.com/middleware/method-override) | Override request method

##### [Learn More](https://echo.labstack.com/middleware/overview)

#### Third-party Middleware

Middleware | Description
:--- | :---
[echoperm](https://github.com/xyproto/echoperm) | Keeping track of users, login states and permissions.
[echopprof](https://github.com/mtojek/echopprof) | Adapt net/http/pprof to labstack/echo.

### Next

- Head over to [guide](https://echo.labstack.com/guide/installation)
- Browse [recipes](https://echo.labstack.com/recipes/hello-world)

### Need help?

- [Hop on to chat](https://gitter.im/labstack/echo)
- [Open an issue](https://github.com/labstack/echo/issues/new)

## Support Us

- :star: the project
- [Donate](https://echo.labstack.com/support-echo)
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
