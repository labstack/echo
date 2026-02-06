// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/labstack/echo/v5"
)

// StaticConfig defines the config for Static middleware.
type StaticConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Root directory from where the static content is served (relative to given Filesystem).
	// `Root: "."` means root folder from Filesystem.
	// Required.
	Root string

	// Filesystem provides access to the static content.
	// Optional. Defaults to echo.Filesystem (serves files from `.` folder where executable is started)
	Filesystem fs.FS

	// Index file for serving a directory.
	// Optional. Default value "index.html".
	Index string

	// Enable HTML5 mode by forwarding all not-found requests to root so that
	// SPA (single-page application) can handle the routing.
	// Optional. Default value false.
	HTML5 bool

	// Enable directory browsing.
	// Optional. Default value false.
	Browse bool

	// Enable ignoring of the base of the URL path.
	// Example: when assigning a static middleware to a non root path group,
	// the filesystem path is not doubled
	// Optional. Default value false.
	IgnoreBase bool

	// DisablePathUnescaping disables path parameter (param: *) unescaping. This is useful when router is set to unescape
	// all parameter and doing it again in this middleware would corrupt filename that is requested.
	DisablePathUnescaping bool

	// DirectoryListTemplate is template to list directory contents
	// Optional. Default to `directoryListHTMLTemplate` constant below.
	DirectoryListTemplate string
}

const directoryListHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="ie=edge">
  <title>{{ .Name }}</title>
  <style>
    body {
			font-family: Menlo, Consolas, monospace;
			padding: 48px;
		}
		header {
			padding: 4px 16px;
			font-size: 24px;
		}
    ul {
			list-style-type: none;
			margin: 0;
    	padding: 20px 0 0 0;
			display: flex;
			flex-wrap: wrap;
    }
    li {
			width: 300px;
			padding: 16px;
		}
		li a {
			display: block;
			overflow: hidden;
			white-space: nowrap;
			text-overflow: ellipsis;
			text-decoration: none;
			transition: opacity 0.25s;
		}
		li span {
			color: #707070;
			font-size: 12px;
		}
		li a:hover {
			opacity: 0.50;
		}
		.dir {
			color: #E91E63;
		}
		.file {
			color: #673AB7;
		}
  </style>
</head>
<body>
	<header>
		{{ .Name }}
	</header>
	<ul>
		{{ range .Files }}
		<li>
		{{ if .Dir }}
			{{ $name := print .Name "/" }}
			<a class="dir" href="{{ $name }}">{{ $name }}</a>
			{{ else }}
			<a class="file" href="{{ .Name }}">{{ .Name }}</a>
			<span>{{ .Size }}</span>
		{{ end }}
		</li>
		{{ end }}
  </ul>
</body>
</html>
`

// DefaultStaticConfig is the default Static middleware config.
var DefaultStaticConfig = StaticConfig{
	Skipper: DefaultSkipper,
	Index:   "index.html",
}

// Static returns a Static middleware to serves static content from the provided root directory.
func Static(root string) echo.MiddlewareFunc {
	c := DefaultStaticConfig
	c.Root = root
	return StaticWithConfig(c)
}

// StaticWithConfig returns a Static middleware to serves static content or panics on invalid configuration.
func StaticWithConfig(config StaticConfig) echo.MiddlewareFunc {
	return toMiddlewareOrPanic(config)
}

// ToMiddleware converts StaticConfig to middleware or returns an error for invalid configuration
func (config StaticConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	// Defaults
	if config.Root == "" {
		config.Root = "." // For security we want to restrict to CWD.
	} else {
		config.Root = path.Clean(config.Root) // fs.Open is very picky about ``, `.`, `..` in paths, so remove some of them up.
	}

	if config.Skipper == nil {
		config.Skipper = DefaultStaticConfig.Skipper
	}
	if config.Index == "" {
		config.Index = DefaultStaticConfig.Index
	}
	if config.DirectoryListTemplate == "" {
		config.DirectoryListTemplate = directoryListHTMLTemplate
	}

	dirListTemplate, tErr := template.New("index").Parse(config.DirectoryListTemplate)
	if tErr != nil {
		return nil, fmt.Errorf("echo static middleware directory list template parsing error: %w", tErr)
	}

	var once *sync.Once
	var fsErr error
	currentFS := config.Filesystem
	if config.Filesystem == nil {
		once = &sync.Once{}
	} else if config.Root != "." {
		tmpFs, fErr := fs.Sub(config.Filesystem, path.Join(".", config.Root))
		if fErr != nil {
			return nil, fmt.Errorf("static middleware failed to create sub-filesystem from config.Root, error: %w", fErr)
		}
		currentFS = tmpFs
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			p := c.Request().URL.Path
			pathUnescape := true
			if strings.HasSuffix(c.Path(), "*") { // When serving from a group, e.g. `/static*`.
				p = c.Param("*")
				pathUnescape = !config.DisablePathUnescaping // because router could already do PathUnescape
			}
			if pathUnescape {
				p, err = url.PathUnescape(p)
				if err != nil {
					return err
				}
			}
			// Security: We use path.Clean() (not filepath.Clean()) because:
			// 1. HTTP URLs always use forward slashes, regardless of server OS
			// 2. path.Clean() provides platform-independent behavior for URL paths
			// 3. The "/" prefix forces absolute path interpretation, removing ".." components
			// 4. Backslashes are treated as literal characters (not path separators), preventing traversal
			// See static_windows.go for Go 1.20+ filepath.Clean compatibility notes
			filePath := path.Clean("./" + p)

			if config.IgnoreBase {
				routePath := path.Base(strings.TrimRight(c.Path(), "/*"))
				baseURLPath := path.Base(p)
				if baseURLPath == routePath {
					i := strings.LastIndex(filePath, routePath)
					filePath = filePath[:i] + strings.Replace(filePath[i:], routePath, "", 1)
				}
			}

			if once != nil {
				once.Do(func() {
					if tmp, tmpErr := fs.Sub(c.Echo().Filesystem, config.Root); tmpErr != nil {
						fsErr = fmt.Errorf("static middleware failed to create sub-filesystem: %w", tmpErr)
					} else {
						currentFS = tmp
					}
				})
				if fsErr != nil {
					return fsErr
				}
			}

			file, err := currentFS.Open(filePath)
			if err != nil {
				if !isIgnorableOpenFileError(err) {
					return err
				}
				// file with that path did not exist, so we continue down in middleware/handler chain, hoping that we end up in
				// handler that is meant to handle this request
				err = next(c)
				if err == nil {
					return nil
				}

				var he echo.HTTPStatusCoder
				if !(errors.As(err, &he) && config.HTML5 && he.StatusCode() == http.StatusNotFound) {
					return err
				}
				// is case HTML5 mode is enabled + echo 404 we serve index to the client
				file, err = currentFS.Open(config.Index)
				if err != nil {
					return err
				}
			}

			defer file.Close()

			info, err := file.Stat()
			if err != nil {
				return err
			}

			if info.IsDir() {
				index, err := currentFS.Open(path.Join(filePath, config.Index))
				if err != nil {
					if config.Browse {
						return listDir(dirListTemplate, filePath, currentFS, c.Response())
					}

					return next(c)
				}

				defer index.Close()

				info, err = index.Stat()
				if err != nil {
					return err
				}

				return serveFile(c, index, info)
			}

			return serveFile(c, file, info)
		}
	}, nil
}

func serveFile(c *echo.Context, file fs.File, info os.FileInfo) error {
	ff, ok := file.(io.ReadSeeker)
	if !ok {
		return errors.New("file does not implement io.ReadSeeker")
	}
	http.ServeContent(c.Response(), c.Request(), info.Name(), info.ModTime(), ff)
	return nil
}

func listDir(t *template.Template, pathInFs string, filesystem fs.FS, res http.ResponseWriter) error {
	files, err := fs.ReadDir(filesystem, pathInFs)
	if err != nil {
		return fmt.Errorf("static middleware failed to read directory for listing: %w", err)
	}

	// Create directory index
	res.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	data := struct {
		Name  string
		Files []any
	}{
		Name: pathInFs,
	}

	for _, f := range files {
		var size int64
		if !f.IsDir() {
			info, err := f.Info()
			if err != nil {
				return fmt.Errorf("static middleware failed to get file info for listing: %w", err)
			}
			size = info.Size()
		}

		data.Files = append(data.Files, struct {
			Name string
			Dir  bool
			Size string
		}{f.Name(), f.IsDir(), format(size)})
	}

	return t.Execute(res, data)
}

// format formats bytes integer to human readable string.
// For example, 31323 bytes will return 30.59KB.
func format(b int64) string {
	multiple := ""
	value := float64(b)

	switch {
	case b >= EB:
		value /= float64(EB)
		multiple = "EB"
	case b >= PB:
		value /= float64(PB)
		multiple = "PB"
	case b >= TB:
		value /= float64(TB)
		multiple = "TB"
	case b >= GB:
		value /= float64(GB)
		multiple = "GB"
	case b >= MB:
		value /= float64(MB)
		multiple = "MB"
	case b >= KB:
		value /= float64(KB)
		multiple = "KB"
	case b == 0:
		return "0"
	default:
		return strconv.FormatInt(b, 10) + "B"
	}

	return fmt.Sprintf("%.2f%s", value, multiple)
}
