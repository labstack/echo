package middleware

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
)

type (
	// StaticConfig defines the config for Static middleware.
	StaticConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Root directory from where the static content is served.
		// Required.
		Root string `json:"root"`

		// Index file for serving a directory.
		// Optional. Default value "index.html".
		Index string `json:"index"`

		// Enable HTML5 mode by forwarding all not-found requests to root so that
		// SPA (single-page application) can handle the routing.
		// Optional. Default value false.
		HTML5 bool `json:"html5"`

		// Enable directory browsing.
		// Optional. Default value false.
		Browse bool `json:"browse"`
	}
)

var (
	// DefaultStaticConfig is the default Static middleware config.
	DefaultStaticConfig = StaticConfig{
		Skipper: DefaultSkipper,
		Index:   "index.html",
	}
)

// Static returns a Static middleware to serves static content from the provided
// root directory.
func Static(root string) echo.MiddlewareFunc {
	c := DefaultStaticConfig
	c.Root = root
	return StaticWithConfig(c)
}

// StaticWithConfig returns a Static middleware with config.
// See `Static()`.
func StaticWithConfig(config StaticConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Root == "" {
		config.Root = "." // For security we want to restrict to CWD.
	}
	if config.Skipper == nil {
		config.Skipper = DefaultStaticConfig.Skipper
	}
	if config.Index == "" {
		config.Index = DefaultStaticConfig.Index
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			p := c.Request().URL.Path
			if strings.HasSuffix(c.Path(), "*") { // When serving from a group, e.g. `/static*`.
				p = c.Param("*")
			}
			p, err = echo.PathUnescape(p)
			if err != nil {
				return
			}
			name := filepath.Join(config.Root, path.Clean("/"+p)) // "/"+ for security

			fi, err := os.Stat(name)
			if err != nil {
				if os.IsNotExist(err) {
					if err = next(c); err != nil {
						if he, ok := err.(*echo.HTTPError); ok {
							if config.HTML5 && he.Code == http.StatusNotFound {
								return c.File(filepath.Join(config.Root, config.Index))
							}
						}
						return
					}
				}
				return
			}

			if fi.IsDir() {
				index := filepath.Join(name, config.Index)
				fi, err = os.Stat(index)

				if err != nil {
					if config.Browse {
						return listDir(name, c.Response())
					}
					if os.IsNotExist(err) {
						return next(c)
					}
					return
				}

				return c.File(index)
			}

			return c.File(name)
		}
	}
}

func listDir(name string, res *echo.Response) (err error) {
	dir, err := os.Open(name)
	if err != nil {
		return
	}
	dirs, err := dir.Readdir(-1)
	if err != nil {
		return
	}

	// Create a directory index
	res.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	if _, err = fmt.Fprintf(res, "<pre>\n"); err != nil {
		return
	}
	for _, d := range dirs {
		name := d.Name()
		color := "#212121"
		if d.IsDir() {
			color = "#e91e63"
			name += "/"
		}
		if _, err = fmt.Fprintf(res, "<a href=\"%s\" style=\"color: %s;\">%s</a>\n", name, color, name); err != nil {
			return
		}
	}
	_, err = fmt.Fprintf(res, "</pre>\n")
	return
}
