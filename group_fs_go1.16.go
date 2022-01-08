//go:build go1.16
// +build go1.16

package echo

import (
	"fmt"
	"io/fs"
	"net/http"
)

// Static implements `Echo#Static()` for sub-routes within the Group.
func (g *Group) Static(pathPrefix, root string) {
	subFs, err := subFS(g.echo.Filesystem, root)
	if err != nil {
		// happens when `root` contains invalid path according to `fs.ValidPath` rules and we are unable to create FS
		panic(fmt.Errorf("invalid root given to group.Static, err %w", err))
	}
	g.StaticFS(pathPrefix, subFs)
}

// StaticFS implements `Echo#StaticFS()` for sub-routes within the Group.
func (g *Group) StaticFS(pathPrefix string, fileSystem fs.FS) {
	g.Add(
		http.MethodGet,
		pathPrefix+"*",
		StaticDirectoryHandler(fileSystem, false),
	)
}

// FileFS implements `Echo#FileFS()` for sub-routes within the Group.
func (g *Group) FileFS(path, file string, filesystem fs.FS, m ...MiddlewareFunc) *Route {
	return g.GET(path, StaticFileHandler(file, filesystem), m...)
}
