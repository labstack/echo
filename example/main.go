package main

import (
	"net/http"

	"github.com/labstack/bolt"
	mw "github.com/labstack/bolt/middleware"
	"github.com/rs/cors"
	"github.com/thoas/stats"
)

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var users map[string]*user

func init() {
	users = map[string]*user{
		"1": &user{
			ID:   "1",
			Name: "Wreck-It Ralph",
		},
	}
}

func createUser(c *bolt.Context) {
	u := new(user)
	if c.Bind(u) {
		users[u.ID] = u
		c.JSON(http.StatusOK, u)
	}
}

func getUsers(c *bolt.Context) {
	c.JSON(http.StatusOK, users)
}

func getUser(c *bolt.Context) {
	c.JSON(http.StatusOK, users[c.P(0)])
}

func main() {
	b := bolt.New()

	//*************************//
	//   Built-in middleware   //
	//*************************//
	b.Use(mw.Logger)

	//****************************//
	//   Third-party middleware   //
	//****************************//
	// https://github.com/rs/cors
	b.Use(cors.Default().Handler)

	// https://github.com/thoas/stats
	s := stats.New()
	b.Use(s.Handler)
	// Route
	b.Get("/stats", func(c *bolt.Context) {
		c.JSON(200, s.Data())
	})

	// Serve index file
	b.Index("public/index.html")

	// Serve static files
	b.Static("/js", "public/js")

	//************//
	//   Routes   //
	//************//
	b.Post("/users", createUser)
	b.Get("/users", getUsers)
	b.Get("/users/:id", getUser)

	// Start server
	b.Run(":8080")
}
