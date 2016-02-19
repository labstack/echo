package middleware

import (
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/labstack/echo"
)

type (
	StaticOptions struct {
		Root   string `json:"root"`
		Index  string `json:"index"`
		Browse bool   `json:"browse"`
	}
)

func Static(root string, options ...*StaticOptions) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		// Default options
		opts := &StaticOptions{Index: "index.html"}
		if len(options) > 0 {
			opts = options[0]
		}

		return echo.HandlerFunc(func(c echo.Context) error {
			fs := http.Dir(root)
			file := c.Request().URI()
			f, err := fs.Open(file)
			if err != nil {
				return next.Handle(c)
			}
			defer f.Close()

			fi, err := f.Stat()
			if err != nil {
				return err
			}

			if fi.IsDir() {
				/* NOTE:
				Not checking the Last-Modified header as it caches the response `304` when
				changing differnt directories for the same path.
				*/
				d := f

				// Index file
				file = path.Join(file, opts.Index)
				f, err = fs.Open(file)
				if err != nil {
					if opts.Browse {
						dirs, err := d.Readdir(-1)
						if err != nil {
							return err
						}

						// Create a directory index
						res := c.Response()
						res.Header().Set(echo.ContentType, echo.TextHTMLCharsetUTF8)
						if _, err = fmt.Fprintf(res, "<pre>\n"); err != nil {
							return err
						}
						for _, d := range dirs {
							name := d.Name()
							color := "#212121"
							if d.IsDir() {
								color = "#e91e63"
								name += "/"
							}
							if _, err = fmt.Fprintf(res, "<a href=\"%s\" style=\"color: %s;\">%s</a>\n", name, color, name); err != nil {
								return err
							}
						}
						_, err = fmt.Fprintf(res, "</pre>\n")
						return err
					}
					return next.Handle(c)
				}
				fi, _ = f.Stat() // Index file stat
			}
			c.Response().WriteHeader(http.StatusOK)
			io.Copy(c.Response(), f)
			return nil
			// TODO:
			// http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
		})
	}
}

// Favicon serves the default favicon - GET /favicon.ico.
func Favicon() echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
}
