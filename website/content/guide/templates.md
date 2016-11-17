+++
title = "Templates"
description = "How to use templates in Echo"
[menu.side]
  name = "Templates"
  parent = "guide"
  weight = 3
+++

## Template Rendering

`Context#Render(code int, name string, data interface{}) error` renders a template
with data and sends a text/html response with status code. Templates can be registered
using `Echo.SetRenderer()`, allowing us to use any template engine.

Example below shows how to use Go `html/template`:

1. Implement `echo.Renderer` interface

    ```go
    type Template struct {
        templates *template.Template
    }

    func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    	return t.templates.ExecuteTemplate(w, name, data)
    }
    ```

2. Pre-compile templates

    `public/views/hello.html`

    ```html
    {{define "hello"}}Hello, {{.}}!{{end}}
    ```

    ```go
    t := &Template{
        templates: template.Must(template.ParseGlob("public/views/*.html")),
    }
    ```

3. Register templates

    ```go
    e := echo.New()
    e.Renderer = t
    e.GET("/hello", Hello)
    ```

4. Render a template inside your handler

    ```go
    func Hello(c echo.Context) error {
    	return c.Render(http.StatusOK, "hello", "World")
    }
    ```
