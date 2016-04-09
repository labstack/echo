package echo

import (
	"fmt"
	"net/http"
	"path"
)

type (
	// StaticConfig defines the config for static handler.
	StaticConfig struct {
		// Root is the directory from where the static content is served.
		// Required.
		Root string `json:"root"`

		// Index is the list of index files to be searched and used when serving
		// a directory.
		// Optional with default value as []string{"index.html"}.
		Index []string `json:"index"`

		// Browse is a flag to enable/disable directory browsing.
		// Optional with default value as false.
		Browse bool `json:"browse"`
	}
)

var (
	// DefaultStaticConfig is the default static handler config.
	DefaultStaticConfig = StaticConfig{
		Index:  []string{"index.html"},
		Browse: false,
	}
)

// Static returns a static handler to serves static content from the provided
// root directory.
func Static(root string) HandlerFunc {
	c := DefaultStaticConfig
	c.Root = root
	return StaticWithConfig(c)
}

// StaticWithConfig returns a static handler from config.
// See `Static()`.
func StaticWithConfig(config StaticConfig) HandlerFunc {
	// Defaults
	if config.Index == nil {
		config.Index = DefaultStaticConfig.Index
	}

	return func(c Context) error {
		fs := http.Dir(config.Root)
		file := path.Clean(c.P(0))
		f, err := fs.Open(file)
		if err != nil {
			return ErrNotFound
		}
		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			return ErrNotFound
		}

		if fi.IsDir() {
			/* NOTE:
			Not checking the Last-Modified header as it caches the response `304` when
			changing different directories for the same path.
			*/
			d := f

			// Index file
			// TODO: search all files
			file = path.Join(file, config.Index[0])
			f, err = fs.Open(file)
			if err != nil {
				return ErrNotFound
			}
			if config.Browse {
				dirs, err := d.Readdir(-1)
				if err != nil {
					return err
				}

				// Create a directory index
				rs := c.Response()
				rs.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
				if _, err = fmt.Fprintf(rs, "<pre>\n"); err != nil {
					return err
				}
				for _, d := range dirs {
					name := d.Name()
					color := "#212121"
					if d.IsDir() {
						color = "#e91e63"
						name += "/"
					}
					if _, err = fmt.Fprintf(rs, "<a href=\"%s\" style=\"color: %s;\">%s</a>\n", name, color, name); err != nil {
						return err
					}
				}
				_, err = fmt.Fprintf(rs, "</pre>\n")
				return err
			}
			if fi, err = f.Stat(); err != nil { // Index file
				return ErrNotFound
			}
		}
		return c.ServeContent(f, fi.Name(), fi.ModTime())
	}
}
