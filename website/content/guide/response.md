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

### File

`func (c *context) File(file string) error`

Sends a response with the content of the file.

### Attachment

`Context#Attachment(file string) error`

Sends a response as file attachment, prompting client to save the file.

### Static Files

`Echo#Use(middleware.Static(root string))`

Serves static files from the provided `root` directory.

`Echo#Static(prefix, root string)`

Serves files from provided `root` directory for `/<prefix>*` HTTP path.

`Echo#File(path, file string)`

Serves provided `file` for `/<path>` HTTP path.

*Examples*

- Serving static files with no prefix `e.Use(middleware.Static("public"))`
- Serving static files with a prefix `e.Static("/static", "assets")`
- Serving an index page `e.File("/", "public/index.html")`
- Serving a favicon `e.File("/favicon.ico", "images/facicon.ico")`
