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
		Age  int
	}
)

var (
	users = map[int]user{}
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
	users[u.ID] = *u
	seq++
	return c.JSON(http.StatusCreated, u)
}

func getUser(c *echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	return c.JSON(http.StatusOK, users[id])
}

func updateUser(c *echo.Context) error {
	// id, _ := strconv.Atoi(c.Param("id"))
	// users[id]
	return c.NoContent(http.StatusNoContent)
}

func deleteUser(c *echo.Context) error {
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
	e.Put("/users/:id", updateUser)
	e.Delete("/users/:id", deleteUser)

	// Start server
	e.Run(":8080")
}
