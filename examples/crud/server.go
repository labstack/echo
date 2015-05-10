package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
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

func createUser(c *echo.Context) *echo.HTTPError {
	u := &user{
		ID: seq,
	}
	if he := c.Bind(u); he != nil {
		return he
	}
	users[u.ID] = u
	seq++
	return c.JSON(http.StatusCreated, u)
}

func getUser(c *echo.Context) *echo.HTTPError {
	id, _ := strconv.Atoi(c.Param("id"))
	return c.JSON(http.StatusOK, users[id])
}

func updateUser(c *echo.Context) *echo.HTTPError {
	u := new(user)
	if he := c.Bind(u); he != nil {
		return he
	}
	id, _ := strconv.Atoi(c.Param("id"))
	users[id].Name = u.Name
	return c.JSON(http.StatusOK, users[id])
}

func deleteUser(c *echo.Context) *echo.HTTPError {
	id, _ := strconv.Atoi(c.Param("id"))
	delete(users, id)
	return c.NoContent(http.StatusNoContent)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(mw.Logger)

	// Routes
	e.Post("/users", createUser)
	e.Get("/users/:id", getUser)
	e.Patch("/users/:id", updateUser)
	e.Delete("/users/:id", deleteUser)

	// Start server
	e.Run(":1323")
}
