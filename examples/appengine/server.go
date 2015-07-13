package main

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/rs/cors"
	"github.com/thoas/stats"
)

type (
	user struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)

var (
	e     = wrapMux()
	users map[string]user
)

// some middleware / handlers that we want to add first for both platforms
func wrapMux() *echo.Echo {

	// this creates the echo mux for the platform
	e := createMux()

	//------------------------
	// Third-party middleware
	//------------------------

	// https://github.com/rs/cors
	e.Use(cors.Default().Handler)

	// https://github.com/thoas/stats
	s := stats.New()
	e.Use(s.Handler)
	// Route
	e.Get("/stats", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, s.Data())
	})

	return e
}

func init() {
	users = map[string]user{
		"1": user{
			ID:   "1",
			Name: "Wreck-It Ralph",
		},
	}
}
