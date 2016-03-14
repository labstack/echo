package middleware

import (
	"fmt"
	"net/http"
	"path"

	"github.com/labstack/echo"
)

type (
	StaticConfig struct {
		Root   string `json:"root"`
		Index  string `json:"index"`
		Browse bool   `json:"browse"`
	}
)

var (
	DefaultStaticConfig = StaticConfig{
		Index:  "index.html",
		Browse: false,
	}
)

func Static(root string) echo.MiddlewareFunc {
	c := DefaultStaticConfig
	c.Root = root
	return StaticFromConfig(c)
}

func StaticFromConfig(config StaticConfig) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			fs := http.Dir(config.Root)
			file := path.Clean(c.Request().URL().Path())
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
				file = path.Join(file, config.Index)
				f, err = fs.Open(file)
				if err != nil {
					if config.Browse {
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
			return echo.ServeContent(c.Request(), c.Response(), f, fi)
		})
	}
}
