---
title: Response
menu:
  side:
    parent: guide
    weight: 6
---

### Template

```go
Context.Render(code int, name string, data interface{}) error
```
Renders a template with data and sends a text/html response with status code. Templates
can be registered using `Echo.SetRenderer()`, allowing us to use any template engine.

Below is an example using Go `html/template`

- Implement `echo.Render` interface

```go
type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}) error {
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
func Hello(c *echo.Context) error {
	return c.Render(http.StatusOK, "hello", "World")
}
```

### JSON

```go
Context.JSON(code int, v interface{}) error
```

Sends a JSON HTTP response with status code.

### JSONP

```go
Context.JSONP(code int, callback string, i interface{}) error
```

Sends a JSONP HTTP response with status code. It uses `callback` to construct the
JSONP payload.

### XML

```go
Context.XML(code int, v interface{}) error
```

Sends an XML HTTP response with status code.

### HTML

```go
Context.HTML(code int, html string) error
```

Sends an HTML response with status code.

### String

```go
Context.String(code int, s string) error
```

Sends a string response with status code.

### File

```go
Context.File(path, name string, attachment bool) error
```

File sends a response with the content of the file. If `attachment` is set
to `true`, the client is prompted to save the file with provided `name`,
name can be empty, in that case name of the file is used.

### Static files

`Echo.Static(path, root string)` serves static files. For example, code below serves
files from directory `public/scripts` for any request path starting with `/scripts/`.

```go
e.Static("/scripts/", "public/scripts")
```

### Serving a file

`Echo.ServeFile(path, file string)` serves a file. For example, code below serves
file `welcome.html` for request path `/welcome`.

```go
e.ServeFile("/welcome", "welcome.html")
```

### Serving an index file

`Echo.Index(file string)` serves root index page - `GET /`. For example, code below
serves root index page from file `public/index.html`.

```go
e.Index("public/index.html")
```

### Serving favicon

`Echo.Favicon(file string)` serves default favicon - `GET /favicon.ico`. For example,
code below serves favicon from file `public/favicon.ico`.

```go
e.Favicon("public/favicon.ico")
```
