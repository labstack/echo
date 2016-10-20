+++
title = "Index"
+++

# Fast and unfancy HTTP server framework for Go (Golang).

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

<img style="width: 100%;" src="https://i.imgur.com/fZVnK52.png" alt="Performance">

## Quick Start

### Installation

```sh
$ go get -u github.com/labstack/echo
```

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
[BodyLimit]({{< ref "middleware/body-limit.md">}}) | Limit request body
[Logger]({{< ref "middleware/logger.md">}}) | Log HTTP requests
[Recover]({{< ref "middleware/recover.md">}}) | Recover from panics
[Gzip]({{< ref "middleware/gzip.md">}}) | Send gzip HTTP response
[BasicAuth]({{< ref "middleware/basic-auth.md">}}) | HTTP basic authentication
[JWTAuth]({{< ref "middleware/jwt.md">}}) | JWT authentication
[Secure]({{< ref "middleware/secure.md">}}) | Protection against attacks
[CORS]({{< ref "middleware/cors.md">}}) | Cross-Origin Resource Sharing
[CSRF]({{< ref "middleware/csrf.md">}}) | Cross-Site Request Forgery
[Static]({{< ref "middleware/static.md">}}) | Serve static files
[HTTPSRedirect]({{< ref "middleware/redirect.md#httpsredirect-middleware">}}) | Redirect HTTP requests to HTTPS
[HTTPSWWWRedirect]({{< ref "middleware/redirect.md#httpswwwredirect-middleware">}}) | Redirect HTTP requests to WWW HTTPS
[WWWRedirect]({{< ref "middleware/redirect.md#wwwredirect-middleware">}}) | Redirect non WWW requests to WWW
[NonWWWRedirect]({{< ref "middleware/redirect.md#nonwwwredirect-middleware">}}) | Redirect WWW requests to non WWW
[AddTrailingSlash]({{< ref "middleware/trailing-slash.md#addtrailingslash-middleware">}}) | Add trailing slash to the request URI
[RemoveTrailingSlash]({{< ref "middleware/trailing-slash.md#removetrailingslash-middleware">}}) | Remove trailing slash from the request URI
[MethodOverride]({{< ref "middleware/method-override.md">}}) | Override request method

#### Third-party Middleware

Middleware | Description
:--- | :---
[echoperm](https://github.com/xyproto/echoperm) | Keeping track of users, login states and permissions.
[echopprof](https://github.com/mtojek/echopprof) | Adapt net/http/pprof to labstack/echo.

##### [Learn More](https://echo.labstack.com/middleware/overview)

### Next

- Head over to [guide](https://echo.labstack.com/guide/installation)
- Browse [recipes](https://echo.labstack.com/recipes/hello-world)

### Need help?

- [Hop on to chat](https://gitter.im/labstack/echo)
- [Open an issue](https://github.com/labstack/echo/issues/new)

## Support Echo

- ☆ the project
- [Donate](https://echo.labstack.com/support-echo)
- 🌐 spread the word
- [Contribute](#contribute:d680e8a854a7cbad6d490c445cba2eba) to the project

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
