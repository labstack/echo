package main

import (
	"net/http"

	"github.com/labstack/bolt"
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
	b.Post("/users", createUser)
	b.Get("/users", getUsers)
	b.Get("/users/:id", getUser)
	b.Run(":8080")
}
