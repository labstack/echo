# [Echo](http://labstack.github.io/echo) [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/labstack/echo) [![Build Status](http://img.shields.io/travis/labstack/echo.svg?style=flat-square)](https://travis-ci.org/labstack/echo) [![Coverage Status](http://img.shields.io/coveralls/labstack/echo.svg?style=flat-square)](https://coveralls.io/r/labstack/echo)
Echo is a fast HTTP router (zero memory allocation) and micro web framework in Go.

## Features

- Fast HTTP router which smartly prioritize routes.
- Extensible middleware/handler, supports:
	- Middleware
		- `echo.MiddlewareFunc`
		- `func(echo.HandlerFunc) echo.HandlerFunc`
		- `echo.HandlerFunc`
		- `func(*echo.Context) error`
		- `func(http.Handler) http.Handler`
		- `http.Handler`
		- `http.HandlerFunc`
		- `func(http.ResponseWriter, *http.Request)`
	- Handler
		- `echo.HandlerFunc`
		- `func(*echo.Context) error`
		- `http.Handler`
		- `http.HandlerFunc`
		- `func(http.ResponseWriter, *http.Request)`
- Sub routing with groups.
- Handy encoding/decoding functions.
- Serve static files, including index.
- Centralized HTTP error handling.
- Use a customized function to bind request body to a Go type.
- Register a view render so you can use any HTML template engine.

## Benchmark

Based on [julienschmidt/go-http-routing-benchmark] (https://github.com/vishr/go-http-routing-benchmark), April 1, 2015

##### [GitHub API](http://developer.github.com/v3)

> Echo: 42728 ns/op, 0 B/op, 0 allocs/op

```
BenchmarkAce_GithubAll	   	   20000	     65328 ns/op	   13792 B/op	     167 allocs/op
BenchmarkBear_GithubAll	   	   10000	    241852 ns/op	   79952 B/op	     943 allocs/op
BenchmarkBeego_GithubAll	    3000	    458234 ns/op	  146272 B/op	    2092 allocs/op
BenchmarkBone_GithubAll	    	1000	   1923508 ns/op	  648016 B/op	    8119 allocs/op
BenchmarkDenco_GithubAll	   20000	     81294 ns/op	   20224 B/op	     167 allocs/op
BenchmarkEcho_GithubAll	   	   30000	     42728 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_GithubAll	   	   20000	     69373 ns/op	   13792 B/op	     167 allocs/op
BenchmarkGocraftWeb_GithubAll  10000	    370978 ns/op	  133280 B/op	    1889 allocs/op
BenchmarkGoji_GithubAll	    	3000	    542766 ns/op	   56113 B/op	     334 allocs/op
BenchmarkGoJsonRest_GithubAll	5000	    452551 ns/op	  135995 B/op	    2940 allocs/op
BenchmarkGoRestful_GithubAll	 200	   9500204 ns/op	  707604 B/op	    7558 allocs/op
BenchmarkGorillaMux_GithubAll	 200	   6770545 ns/op	  153137 B/op	    1791 allocs/op
BenchmarkHttpRouter_GithubAll  30000	     56097 ns/op	   13792 B/op	     167 allocs/op
BenchmarkHttpTreeMux_GithubAll 10000	    143175 ns/op	   56112 B/op	     334 allocs/op
BenchmarkKocha_GithubAll	   10000	    147959 ns/op	   23304 B/op	     843 allocs/op
BenchmarkMacaron_GithubAll	    2000	    724650 ns/op	  224960 B/op	    2315 allocs/op
BenchmarkMartini_GithubAll	     100	  10926021 ns/op	  237953 B/op	    2686 allocs/op
BenchmarkPat_GithubAll	     	 300	   4525114 ns/op	 1504101 B/op	   32222 allocs/op
BenchmarkRevel_GithubAll	    2000	   1172963 ns/op	  345553 B/op	    5918 allocs/op
BenchmarkRivet_GithubAll	   10000	    249104 ns/op	   84272 B/op	    1079 allocs/op
BenchmarkTango_GithubAll	     300	   4012826 ns/op	 1368581 B/op	   29157 allocs/op
BenchmarkTigerTonic_GithubAll	2000	    975450 ns/op	  241088 B/op	    6052 allocs/op
BenchmarkTraffic_GithubAll	     200	   7540377 ns/op	 2664762 B/op	   22390 allocs/op
BenchmarkVulcan_GithubAll	    5000	    307241 ns/op	   19894 B/op	     609 allocs/op
BenchmarkZeus_GithubAll	        2000	    752907 ns/op	  300688 B/op	    2648 allocs/op
```

## Installation

```sh
$ go get github.com/labstack/echo
```

##[Examples](https://github.com/labstack/echo/tree/master/examples)

###[Hello, World!](https://github.com/labstack/echo/tree/master/examples/hello)

```go
package main

import (
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

// Handler
func hello(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())

	// Routes
	e.Get("/", hello)

	// Start server
	e.Run(":1323")
}
```

###[CRUD](https://github.com/labstack/echo/tree/master/examples/crud)

```go
package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

type (
	user struct {
		ID   int
		Name string
	}
)

var (
	users = map[int]*user{}
	seq   = 1
)

//----------
// Handlers
//----------

func createUser(c *echo.Context) error {
	u := &user{
		ID: seq,
	}
	if err := c.Bind(u); err != nil {
		return err
	}
	users[u.ID] = u
	seq++
	return c.JSON(http.StatusCreated, u)
}

func getUser(c *echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	return c.JSON(http.StatusOK, users[id])
}

func updateUser(c *echo.Context) error {
	u := new(user)
	if err := c.Bind(u); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	users[id].Name = u.Name
	return c.JSON(http.StatusOK, users[id])
}

func deleteUser(c *echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	delete(users, id)
	return c.NoContent(http.StatusNoContent)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())

	// Routes
	e.Post("/users", createUser)
	e.Get("/users/:id", getUser)
	e.Patch("/users/:id", updateUser)
	e.Delete("/users/:id", deleteUser)

	// Start server
	e.Run(":1323")
}
```

###[Website](https://github.com/labstack/echo/tree/master/examples/website)

```go
package main

import (
	"io"
	"net/http"

	"html/template"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/rs/cors"
	"github.com/thoas/stats"
)

type (
	// Template provides HTML template rendering
	Template struct {
		templates *template.Template
	}

	user struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)

var (
	users map[string]user
)

// Render HTML
func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

//----------
// Handlers
//----------

func welcome(c *echo.Context) error {
	return c.Render(http.StatusOK, "welcome", "Joe")
}

func createUser(c *echo.Context) error {
	u := new(user)
	if err := c.Bind(u); err != nil {
		return err
	}
	users[u.ID] = *u
	return c.JSON(http.StatusCreated, u)
}

func getUsers(c *echo.Context) error {
	return c.JSON(http.StatusOK, users)
}

func getUser(c *echo.Context) error {
	return c.JSON(http.StatusOK, users[c.P(0)])
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())

	//------------------------
	// Third-party middleware
	//------------------------

	// https://github.com/rs/cors
	e.Use(cors.Default().Handler)

	// https://github.com/thoas/stats
	s := stats.New()
	e.Use(s.Handler)
	// Route
	e.Get("/stats", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, s.Data())
	})

	// Serve index file
	e.Index("public/index.html")

	// Serve favicon
	e.Favicon("public/favicon.ico")

	// Serve static files
	e.Static("/scripts/", "public/scripts")

	//--------
	// Routes
	//--------

	e.Post("/users", createUser)
	e.Get("/users", getUsers)
	e.Get("/users/:id", getUser)

	//-----------
	// Templates
	//-----------

	t := &Template{
		// Cached templates
		templates: template.Must(template.ParseFiles("public/views/welcome.html")),
	}
	e.SetRenderer(t)
	e.Get("/welcome", welcome)

	//-------
	// Group
	//-------

	// Group with parent middleware
	a := e.Group("/admin")
	a.Use(func(c *echo.Context) error {
		// Security middleware
		return nil
	})
	a.Get("", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Welcome admin!")
	})

	// Group with no parent middleware
	g := e.Group("/files", func(c *echo.Context) error {
		// Security middleware
		return nil
	})
	g.Get("", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Your files!")
	})

	// Start server
	e.Run(":1323")
}

func init() {
	users = map[string]user{
		"1": user{
			ID:   "1",
			Name: "Wreck-It Ralph",
		},
	}
}
```

###[Middleware](https://github.com/labstack/echo/tree/master/examples/middleware)

```go
package main

import (
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

// Handler
func hello(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func main() {
	// Echo instance
	e := echo.New()

	// Debug mode
	e.SetDebug(true)

	//------------
	// Middleware
	//------------

	// Logger
	e.Use(mw.Logger())

	// Recover
	e.Use(mw.Recover())

	// Basic auth
	e.Use(mw.BasicAuth(func(u, p string) bool {
		if u == "joe" && p == "secret" {
			return true
		}
		return false
	}))

	//-------
	// Slash
	//-------

	e.Use(mw.StripTrailingSlash())

	// or

	//	e.Use(mw.RedirectToSlash())

	// Gzip
	e.Use(mw.Gzip())

	// Routes
	e.Get("/", hello)

	// Start server
	e.Run(":1323")
}
```

##[Guide](http://labstack.github.io/echo/guide)

## Contribute

**Use issues for everything**

- Report problems
- Discuss before sending pull request
- Suggest new features
- Improve/fix documentation

## Credits
- [Vishal Rana](https://github.com/vishr) - Author
- [Nitin Rana](https://github.com/nr17) - Consultant
- [Contributors](https://github.com/labstack/echo/graphs/contributors)

## License

[MIT](https://github.com/labstack/echo/blob/master/LICENSE)
