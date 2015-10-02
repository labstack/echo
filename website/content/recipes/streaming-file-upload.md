---
title: Streaming File Upload
menu:
  main:
    parent: recipes
---

- Streaming multipart/form-data file upload
- Multiple form fields and files

## Server

`server.go`

```go
package main

import (
	"io/ioutil"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"io"
	"net/http"
	"os"
)

func upload(c *echo.Context) error {
	mr, err := c.Request().MultipartReader()
	if err != nil {
		return err
	}

	// Read form field `name`
	part, err := mr.NextPart()
	if err != nil {
		return err
	}
	defer part.Close()
	b, err := ioutil.ReadAll(part)
	if err != nil {
		return err
	}
	name := string(b)

	// Read form field `email`
	part, err = mr.NextPart()
	if err != nil {
		return err
	}
	defer part.Close()
	b, err = ioutil.ReadAll(part)
	if err != nil {
		return err
	}
	email := string(b)

	// Read files
	i := 0
	for {
		part, err := mr.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		defer part.Close()

		file, err := os.Create(part.FileName())
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(file, part); err != nil {
			return err
		}
		i++
	}
	return c.String(http.StatusOK, "Thank You! %s <%s>, %d files uploaded successfully.",
		name, email, i)
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

## [Source Code](https://github.com/labstack/echo/blob/master/recipes/streaming-file-upload)
