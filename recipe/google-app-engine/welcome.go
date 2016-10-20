package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo"
)

type (
	Template struct {
		templates *template.Template
	}
)

func init() {
	t := &Template{
		templates: template.Must(template.ParseFiles("templates/welcome.html")),
	}
	e.Renderer = t
	e.GET("/welcome", welcome)
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func welcome(c echo.Context) error {
	return c.Render(http.StatusOK, "welcome", "Joe")
}
