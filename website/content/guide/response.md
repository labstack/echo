---
title: Response
menu:
  side:
    parent: guide
    weight: 7
---

### Template

`Context#Render(code int, name string, data interface{}) error`

Renders a template with data and sends a text/html response with status code. Templates
can be registered using `Echo.SetRenderer()`, allowing us to use any template engine.

Below is an example using Go `html/template`

- Implement `echo.Render` interface

```go
type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
```

- Pre-compile templates

`public/views/hello.html`

```html
{{define "hello"}}Hello, {{.}}!{{end}}
```

```go
t := &Template{
    templates: template.Must(template.ParseGlob("public/views/*.html")),
}
```

- Register templates

```go
e := echo.New()
e.SetRenderer(t)
e.Get("/hello", Hello)
```

- Render template

```go
func Hello(c echo.Context) error {
	return c.Render(http.StatusOK, "hello", "World")
}
```

### JSON

`Context.JSON(code int, v interface{}) error`

Sends a JSON HTTP response with status code.

### JSONP

`Context.JSONP(code int, callback string, i interface{}) error`

Sends a JSONP HTTP response with status code. It uses `callback` to construct the
JSONP payload.

### XML

`Context.XML(code int, v interface{}) error`

Sends an XML HTTP response with status code.

### HTML

`c.HTML(code int, html string) error`

Sends an HTML response with status code.

### String

`Context#String(code int, s string) error`

Sends a string response with status code.

### Attachment

`Context#Attachment(file string) (err error)`

Sends file as an attachment.

### Static files

`Echo#Use(middleware.Static("public"))`

Serves static files from public folder.
