package main

import (
	"io"
	"net/http"

	"html/template"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/rs/cors"
	"github.com/thoas/stats"
)

type (
	// Template provides HTML template rendering
	Template struct {
		templates *template.Template
	}

	user struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)

var (
	users map[string]user
)

// Render HTML
func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

//----------
// Handlers
//----------

func welcome(c *echo.Context) error {
	return c.Render(http.StatusOK, "welcome", "Joe")
}

func createUser(c *echo.Context) error {
	u := new(user)
	if err := c.Bind(u); err != nil {
		return err
	}
	users[u.ID] = *u
	return c.JSON(http.StatusCreated, u)
}

func getUsers(c *echo.Context) error {
	return c.JSON(http.StatusOK, users)
}

func getUser(c *echo.Context) error {
	return c.JSON(http.StatusOK, users[c.P(0)])
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())
	e.Use(mw.Gzip())

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

	// Serve index file
	e.Index("public/index.html")

	// Serve favicon
	e.Favicon("public/favicon.ico")

	// Serve static files
	e.Static("/scripts", "public/scripts")

	//--------
	// Routes
	//--------

	e.Post("/users", createUser)
	e.Get("/users", getUsers)
	e.Get("/users/:id", getUser)

	//-----------
	// Templates
	//-----------

	t := &Template{
		// Cached templates
		templates: template.Must(template.ParseFiles("public/views/welcome.html")),
	}
	e.SetRenderer(t)
	e.Get("/welcome", welcome)

	//-------
	// Group
	//-------

	// Group with parent middleware
	a := e.Group("/admin")
	a.Use(func(c *echo.Context) error {
		// Security middleware
		return nil
	})
	a.Get("", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Welcome admin!")
	})

	// Group with no parent middleware
	g := e.Group("/files", func(c *echo.Context) error {
		// Security middleware
		return nil
	})
	g.Get("", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Your files!")
	})

	// Start server
	e.Run(":1323")
}

func init() {
	users = map[string]user{
		"1": user{
			ID:   "1",
			Name: "Wreck-It Ralph",
		},
	}
}
