package main

import (
	"net/http"

	"github.com/labstack/bolt"
)

type user struct {
	Id   string
	Name string
}

var users map[string]*user

func init() {
	users = map[string]*user{
		"1": &user{
			Id:   "1",
			Name: "Wreck-It Ralph",
		},
	}
}

func createUser(c *bolt.Context) {
}

func getUsers(c *bolt.Context) {
	c.Render(http.StatusOK, bolt.FMT_JSON, users)
}

func getUser(c *bolt.Context) {
	c.Render(http.StatusOK, bolt.FMT_JSON, users[c.P(0)])
}

func main() {
	b := bolt.New()
	b.Get("/users", getUsers)
	b.Get("/users/:id", getUser)
	// go b.RunHttp(":8080")
	// go b.RunWebSocket(":8081")
	b.RunTcp(":8082")
}
