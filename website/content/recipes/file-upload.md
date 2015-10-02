---
title: File Upload
menu:
  main:
    parent: recipes 
---

- Multipart/form-data file upload
- Multiple form fields and files

Use `req.ParseMultipartForm(16 << 20)` for manually parsing multipart form. It gives
us an option to specify the maximum memory used while parsing the request body.

## Server

`server.go`

```go
package main

import (
	"io"
	"os"

	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

func upload(c *echo.Context) error {
	req := c.Request()
	req.ParseMultipartForm(16 << 20) // Max memory 16 MiB

	// Read form fields
	name := c.Form("name")
	email := c.Form("email")

	// Read files
	files := req.MultipartForm.File["files"]
	for _, f := range files {
		// Source file
		src, err := f.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Destination file
		dst, err := os.Create(f.Filename)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return err
		}
	}
	return c.String(http.StatusOK, "Thank You! %s <%s>, %d files uploaded successfully.",
		name, email, len(files))
}

func main() {
	e := echo.New()
	e.Use(mw.Logger())
	e.Use(mw.Recover())

	e.Static("/", "public")
	e.Post("/upload", upload)

	e.Run(":1323")
}
```

## Client

`index.html`

```html
<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>File Upload</title>
</head>
<body>
    <h1>Upload Files</h1>
    <form action="/upload" method="post" enctype="multipart/form-data">
        Name: <input type="text" name="name"><br>
        Email: <input type="email" name="email"><br>
        Files: <input type="file" name="files" multiple><br><br>
        <input type="submit" value="Submit">
    </form>
</body>
</html>

```

## [Source Code](https://github.com/labstack/echo/blob/master/recipes/file-upload)
