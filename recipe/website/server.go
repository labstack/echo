package main

import (
	"io"
	"net/http"

	"html/template"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
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
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

//----------
// Handlers
//----------

func welcome(c echo.Context) error {
	return c.Render(http.StatusOK, "welcome", "Joe")
}

func createUser(c echo.Context) error {
	u := new(user)
	if err := c.Bind(u); err != nil {
		return err
	}
	users[u.ID] = *u
	return c.JSON(http.StatusCreated, u)
}

func getUsers(c echo.Context) error {
	return c.JSON(http.StatusOK, users)
}

func getUser(c echo.Context) error {
	return c.JSON(http.StatusOK, users[c.P(0)])
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.Static("public"))

	//------------------------
	// Third-party middleware
	//------------------------

	// https://github.com/rs/cors
	// e.Use(standard.WrapMiddleware(cors.Default().Handler))

	// https://github.com/thoas/stats
	s := stats.New()
	e.Use(standard.WrapMiddleware(s.Handler))
	e.GET("/stats", echo.HandlerFunc(func(c echo.Context) error {
		return c.JSON(222, s.Data())
	}))

	//--------
	// Routes
	//--------

	e.POST("/users", createUser)
	e.GET("/users", getUsers)
	e.GET("/users/:id", getUser)

	//-----------
	// Templates
	//-----------

	t := &Template{
		// Cached templates
		templates: template.Must(template.ParseFiles("public/views/welcome.html")),
	}
	e.SetRenderer(t)
	e.GET("/welcome", welcome)

	//-------
	// Group
	//-------

	a := e.Group("/admin")
	a.Use(echo.WrapMiddleware(echo.HandlerFunc(func(c echo.Context) error {
		// Security middleware
		return nil
	})))
	a.Get("", echo.HandlerFunc(func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome admin!")
	}))

	// Start server
	e.Run(standard.New(":1323"))
}

func init() {
	users = map[string]user{
		"1": user{
			ID:   "1",
			Name: "Wreck-It Ralph",
		},
	}
}
