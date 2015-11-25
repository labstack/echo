package main

import (
	"fmt"
	"io/ioutil"

	"io"
	"net/http"
	"os"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
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
	return c.String(http.StatusOK, fmt.Sprintf("Thank You! %s <%s>, %d files uploaded successfully.",
		name, email, i))
}

func main() {
	e := echo.New()
	e.Use(mw.Logger())
	e.Use(mw.Recover())

	e.Static("/", "public")
	e.Post("/upload", upload)

	e.Run(":1323")
}
