package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo"
)

type (
	// Template provides HTML template rendering
	Template struct {
		templates *template.Template
	}
)

// Render HTML
func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func welcome(c *echo.Context) error {
	return c.Render(http.StatusOK, "welcome", "Joe")
}

func init() {
	//-----------
	// Templates
	//-----------

	t := &Template{
		// Cached templates
		templates: template.Must(template.ParseFiles("public/views/welcome.html")),
	}
	e.SetRenderer(t)
	e.Get("/welcome", welcome)
}
