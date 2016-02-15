package handler

import (
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/labstack/echo"
)

type (
	Static struct {
		Root   string
		Browse bool
		Index  string
	}

	FaviconOptions struct {
	}
)

func (s Static) Handle(c echo.Context) error {
	fs := http.Dir(s.Root)
	file := c.P(0)
	f, err := fs.Open(file)
	if err != nil {
		return echo.ErrNotFound
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
		file = path.Join(file, s.Index)
		f, err = fs.Open(file)
		if err != nil {
			if s.Browse {
				dirs, err := d.Readdir(-1)
				if err != nil {
					return err
				}

				// Create a directory index
				w := c.Response()
				w.Header().Set(echo.ContentType, echo.TextHTMLCharsetUTF8)
				if _, err = fmt.Fprintf(w, "<pre>\n"); err != nil {
					return err
				}
				for _, d := range dirs {
					name := d.Name()
					color := "#212121"
					if d.IsDir() {
						color = "#e91e63"
						name += "/"
					}
					if _, err = fmt.Fprintf(w, "<a href=\"%s\" style=\"color: %s;\">%s</a>\n", name, color, name); err != nil {
						return err
					}
				}
				_, err = fmt.Fprintf(w, "</pre>\n")
				return err
			}
			return echo.ErrNotFound
		}
		fi, _ = f.Stat() // Index file stat
	}
	c.Response().WriteHeader(http.StatusOK)
	io.Copy(c.Response(), f)
	return nil
	// TODO:
	// http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
}

// Favicon serves the default favicon - GET /favicon.ico.
func Favicon(root string, options ...FaviconOptions) echo.MiddlewareFunc {
	return func(h echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			return nil
		})
	}
}
