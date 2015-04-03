package main

import (
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/rs/cors"
	"github.com/thoas/stats"
)

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var users map[string]user

func init() {
	users = map[string]user{
		"1": user{
			ID:   "1",
			Name: "Wreck-It Ralph",
		},
	}
}

func createUser(c *echo.Context) {
	u := new(user)
	if err := c.Bind(u); err == nil {
		users[u.ID] = *u
		if err := c.JSON(http.StatusCreated, u); err == nil {
			// Do something!
		}
		return
	}
	http.Error(c.Response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func getUsers(c *echo.Context) {
	c.JSON(http.StatusOK, users)
}

func getUser(c *echo.Context) {
	c.JSON(http.StatusOK, users[c.P(0)])
}

func main() {
	e := echo.New()

	//*************************//
	//   Built-in middleware   //
	//*************************//
	e.Use(mw.Logger)

	//****************************//
	//   Third-party middleware   //
	//****************************//
	// https://github.com/rs/cors
	e.Use(cors.Default().Handler)

	// https://github.com/thoas/stats
	s := stats.New()
	e.Use(s.Handler)
	// Route
	e.Get("/stats", func(c *echo.Context) {
		c.JSON(200, s.Data())
	})

	// Serve index file
	e.Index("public/index.html")

	// Serve static files
	e.Static("/js", "public/js")

	//************//
	//   Routes   //
	//************//
	e.Post("/users", createUser)
	e.Get("/users", getUsers)
	e.Get("/users/:id", getUser)

	//****************//
	//   Sub router   //
	//****************//
	// Sub - inherits parent middleware
	sub := e.Sub("/sub")
	sub.Use(func(c *echo.Context) { // Middleware
	})
	sub.Get("/home", func(c *echo.Context) {
		c.String(200, "Sub route /sub/welcome")
	})

	// Group - doesn't inherit parent middleware
	grp := e.Group("/group")
	grp.Use(func(c *echo.Context) { // Middleware
	})
	grp.Get("/home", func(c *echo.Context) {
		c.String(200, "Group route /group/welcome")
	})

	// Start server
	e.Run(":8080")
}
