# Echo

Simple and performant web development!

---

## Overview

Echo is a fast HTTP router (zero memory allocation) and micro web framework in Go.

## Features

- Fast HTTP router which smartly resolves conflicting routes.
- Extensible middleware/handler, supports:
	- Middleware
		- `func(*echo.Context)`
		- `func(*echo.Context) error`
		- `func(echo.HandlerFunc) echo.HandlerFunc`
		- `http.Handler`
		- `http.HandlerFunc`
		- `func(http.ResponseWriter, *http.Request)`
		- `func(http.ResponseWriter, *http.Request) error`
	- Handler
		- `func(*echo.Context)`
		- `func(*echo.Context) error`
		- `http.Handler`
		- `http.HandlerFunc`
		- `func(http.ResponseWriter, *http.Request)`
		- `func(http.ResponseWriter, *http.Request) error`
- Sub routing with groups.
- Handy encoding/decoding functions.
- Serve static files, including index.
- Centralized HTTP error handling.
- Customized binder to decode request body to a Go type.
- Customized view render so you can use any templating engine.

## Installation

```go get github.com/labstack/echo```

## Examples

[labstack/echo/example](https://github.com/labstack/echo/tree/master/examples)

> Hello, World!

Create ```server.go``` with the following content:
```go
package main

import (
	"net/http"

	"github.com/labstack/echo"
)

// Handler
func hello(c *echo.Context) {
	c.String(http.StatusOK, "Hello, World!\n")
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(echo.Logger)

	// Routes
	e.Get("/", hello)

	// Start server
	e.Run(":4444")
}
```

```go run server.go``` & browse to ```http://localhost:8080```

> CRUD - Create, read, update and delete.

- Create user
```curl -X POST -H "Content-Type: application/json" -d '{"name":"Joe"}' http://localhost:4444/users```
- Get user
```curl http://localhost:4444/users/1```
- Update user: Change user's name to Sid
```curl -X PATCH -H "Content-Type: application/json" -d '{"name":"Sid"}' http://localhost:4444/users/1```
- Delete user
```curl -X DELETE http://localhost:4444/users/1```


```go
package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

type (
	user struct {
		ID   int
		Name string
	}
)

var (
	users = map[int]*user{}
	seq   = 1
)

//----------
// Handlers
//----------

func createUser(c *echo.Context) error {
	u := &user{
		ID: seq,
	}
	if err := c.Bind(u); err != nil {
		return err
	}
	users[u.ID] = u
	seq++
	return c.JSON(http.StatusCreated, u)
}

func getUser(c *echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	return c.JSON(http.StatusOK, users[id])
}

func updateUser(c *echo.Context) error {
	u := new(user)
	if err := c.Bind(u); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	users[id].Name = u.Name
	return c.JSON(http.StatusOK, users[id])
}

func deleteUser(c *echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	delete(users, id)
	return c.NoContent(http.StatusNoContent)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(echo.Logger)

	// Routes
	e.Post("/users", createUser)
	e.Get("/users/:id", getUser)
	e.Patch("/users/:id", updateUser)
	e.Delete("/users/:id", deleteUser)

	// Start server
	e.Run(":4444")
}
```

## Contribute
- Report issues
- Suggest new features
- Participate in discussion
- Improve documentation

## License

[MIT](https://github.com/labstack/echo/blob/master/LICENSE)
