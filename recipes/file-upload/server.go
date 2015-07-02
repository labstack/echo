package main

import (
	"io"
	"os"

	"net/http"

	"github.com/labstack/echo"
)

func upload(c *echo.Context) error {
	req := c.Request()

	// req.ParseMultipartForm(16 << 20) // Max memory 16 MiB

	// Read form fields
	name := req.FormValue("name")
	email := req.FormValue("email")

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
	e.Index("public/index.html")
	e.Post("/upload", upload)
	e.Run(":1323")
}
