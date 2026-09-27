// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package middleware

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/bytes"
)

// StaticConfig defines the config for Static middleware.
type StaticConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Root directory from where the static content is served.
	// Required.
	Root string `yaml:"root"`

	// Index file for serving a directory.
	// Optional. Default value "index.html".
	Index string `yaml:"index"`

	// Enable HTML5 mode by forwarding all not-found requests to root so that
	// SPA (single-page application) can handle the routing.
	// Optional. Default value false.
	HTML5 bool `yaml:"html5"`

	// Enable directory browsing.
	// Optional. Default value false.
	Browse bool `yaml:"browse"`

	// Enable ignoring of the base of the URL path.
	// Example: when assigning a static middleware to a non root path group,
	// the filesystem path is not doubled
	// Optional. Default value false.
	IgnoreBase bool `yaml:"ignoreBase"`

	// Filesystem provides access to the static content.
	// Optional. Defaults to http.Dir(config.Root)
	Filesystem http.FileSystem `yaml:"-"`

	// EnablePathUnescaping enables unescaping of the request path (or of the wildcard param `*` when the middleware is
	// used on a wildcard route) before the file is looked up.
	// Default false (safe): the path is used in the same form as the router matched it, so encoded characters such as
	// encoded slashes (%2f) are NOT decoded, preventing ACL bypass where /admin%2fprivate.txt bypasses a /admin/* route
	// guard by not matching that route but being decoded to admin/private.txt. As a consequence, file names that the
	// client sends with non-default escaping (e.g. `%2C`, `%40` or lowercase hex like `%c3%a9`) are not found.
	// Set to true only when serving files whose names need such unescaping and you are not relying on route-based
	// ACL guards to restrict access. Paths with ".", ".." or empty segments are never served, also after unescaping.
	//
	// Enabling echo.RouterConfig.UseEscapedPathForMatching makes this field irrelevant and can lead to security issues when
	// using different Routes to exclude some of the files from being served.
	// e.g. if you serve files from directory as such and use different route to exclude some of the files from being served.
	// 0. given folder structure:
	//   public/
	//   public/index.html
	//   public/admin/private.txt
	// 1. share `public/` folder contents from the server root with `e.Static("/", "public")`
	// 2. naively assume that everything under /admin folder is now forbidden
	//       e.GET("/admin/*", func(c *Context) error { return echo.ErrForbidden })
	// Then request to `/assets/../admin%2fprivate.txt` will be served as router does not match it to guarded route.
	EnablePathUnescaping bool `yaml:"enablePathUnescaping"`
}

const html = `
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

// Static returns a Static middleware to serves static content from the provided
// root directory.
//
// Security: when registered with Echo#Use, the middleware runs before route and group middleware, so guards on routes
// or groups (for example authentication on an `/admin` group) do not protect the files it serves. Keep files that
// need protection outside the root directory, or serve them with Echo#Static / Group#Static behind the guard.
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
	if config.Filesystem == nil {
		config.Filesystem = http.Dir(config.Root)
		config.Root = "."
	}

	// Index template
	t, tErr := template.New("index").Parse(html)
	if tErr != nil {
		panic(fmt.Errorf("echo: %w", tErr))
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			// Resolve the file from the same form of the path that the router matched: the escaped path when it
			// differs from the default encoding. Using the decoded path would let `/admin%2Fsecret.txt` reach a file
			// under a guarded `/admin/*` route (GHSA-375p-5qhx-8wq4).
			p := req.URL.Path
			if req.URL.RawPath != "" {
				p = req.URL.RawPath
			}
			// A path with a ".", ".." or empty segment is resolved by path.Clean() to a different file than the path
			// the router matched, so it is not served as a file (GHSA-3pmx-cf9f-34xr).
			unclean := hasUncleanPath(req)
			if strings.HasSuffix(c.Path(), "*") { // When serving from a group, e.g. `/static*`.
				p = c.Param("*")
				unclean = unclean || hasDotOrEmptySegment(p)
			}
			if !unclean && config.EnablePathUnescaping {
				p, err = url.PathUnescape(p)
				if err != nil {
					return
				}
				unclean = hasDotOrEmptySegment(p) // unescaping can create new dot segments, e.g. `%2e%2e`
			}
			// Security: We use path.Clean() (not filepath.Clean()) because:
			// 1. HTTP URLs always use forward slashes, regardless of server OS
			// 2. path.Clean() provides platform-independent behavior for URL paths
			// 3. The "/" prefix forces absolute path interpretation, removing ".." components
			// 4. Backslashes are treated as literal characters (not path separators), preventing traversal
			// See static_windows.go for Go 1.20+ filepath.Clean compatibility notes
			name := path.Join(config.Root, path.Clean("/"+p)) // "/"+ for security

			if config.IgnoreBase {
				routePath := path.Base(strings.TrimRight(c.Path(), "/*"))
				baseURLPath := path.Base(p)
				if baseURLPath == routePath {
					i := strings.LastIndex(name, routePath)
					name = name[:i] + strings.Replace(name[i:], routePath, "", 1)
				}
			}

			var file http.File
			if unclean {
				err = os.ErrNotExist // handle like a missing file, so HTML5 mode can still serve the index
			} else {
				file, err = config.Filesystem.Open(name)
			}
			if err != nil {
				if !isIgnorableOpenFileError(err) {
					return err
				}

				// file with that path did not exist, so we continue down in middleware/handler chain, hoping that we end up in
				// handler that is meant to handle this request
				if err = next(c); err == nil {
					return err
				}

				var he *echo.HTTPError
				if !(errors.As(err, &he) && config.HTML5 && he.Code == http.StatusNotFound) {
					return err
				}

				file, err = config.Filesystem.Open(path.Join(config.Root, config.Index))
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
				index, err := config.Filesystem.Open(path.Join(name, config.Index))
				if err != nil {
					if config.Browse {
						return listDir(t, name, file, c.Response())
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
	}
}

func serveFile(c echo.Context, file http.File, info os.FileInfo) error {
	http.ServeContent(c.Response(), c.Request(), info.Name(), info.ModTime(), file)
	return nil
}

func listDir(t *template.Template, name string, dir http.File, res *echo.Response) (err error) {
	files, err := dir.Readdir(-1)
	if err != nil {
		return
	}

	// Create directory index
	res.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	data := struct {
		Name  string
		Files []interface{}
	}{
		Name: name,
	}
	for _, f := range files {
		data.Files = append(data.Files, struct {
			Name string
			Dir  bool
			Size string
		}{f.Name(), f.IsDir(), bytes.Format(f.Size())})
	}
	return t.Execute(res, data)
}

// hasDotOrEmptySegment reports whether URL path p has a ".", ".." or empty segment. A single leading and a single
// trailing slash are allowed.
// Keep in sync with the copy in echo_fs.go.
func hasDotOrEmptySegment(p string) bool {
	p = strings.TrimPrefix(p, "/")
	p = strings.TrimSuffix(p, "/")
	if p == "" {
		return false
	}
	for segment := range strings.SplitSeq(p, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return true
		}
	}
	return false
}

// hasUncleanPath reports whether the request path, in the form the router matches by default (the escaped path when it
// differs from the default encoding), has a ".", ".." or empty segment. Encoded dots such as `%2E%2E` are not
// segments here; they only act as `..` when path unescaping for static files is enabled.
func hasUncleanPath(req *http.Request) bool {
	if req.URL.RawPath != "" {
		return hasDotOrEmptySegment(req.URL.RawPath)
	}
	return hasDotOrEmptySegment(req.URL.Path)
}
