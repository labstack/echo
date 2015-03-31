# Echo [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/labstack/echo) [![Build Status](http://img.shields.io/travis/fatih/structs.svg?style=flat-square)](https://travis-ci.org/labstack/echo) [![Coverage Status](http://img.shields.io/coveralls/labstack/echo.svg?style=flat-square)](https://coveralls.io/r/labstack/echo)
Echo is a fast HTTP router (zero memory allocation) + micro web framework in Go.

### Features
- Zippy router.
- Extensible middleware/handler, supports:
	- Middleware
		- `func(*echo.Context)`
		- `func(echo.HandlerFunc) echo.HandlerFunc`
		- `func(http.Handler) http.Handler`
		- `http.Handler`
		- `http.HandlerFunc`
		- `func(http.ResponseWriter, *http.Request)`
	- Handler
		- `func(*echo.Context)`
		- `http.Handler`
		- `http.HandlerFunc`
		- `func(http.ResponseWriter, *http.Request)`
- Handy encoding/decoding functions.
- Serve static files, including index.

### Installation
```go get github.com/labstack/echo```

### Usage
[labstack/echo/example](https://github.com/labstack/echo/tree/master/example)

```go
package main

import (
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/rs/cors"
	"github.com/thoas/stats"
)

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var users map[string]user

func init() {
	users = map[string]user{
		"1": user{
			ID:   "1",
			Name: "Wreck-It Ralph",
		},
	}
}

func createUser(c *echo.Context) {
	u := new(user)
	if c.Bind(u) {
		users[u.ID] = *u
		c.JSON(http.StatusCreated, u)
	}
}

func getUsers(c *echo.Context) {
	c.JSON(http.StatusOK, users)
}

func getUser(c *echo.Context) {
	c.JSON(http.StatusOK, users[c.P(0)])
}

func main() {
	e := echo.New()

	//*************************//
	//   Built-in middleware   //
	//*************************//
	e.Use(mw.Logger)

	//****************************//
	//   Third-party middleware   //
	//****************************//
	// https://github.com/rs/cors
	e.Use(cors.Default().Handler)

	// https://github.com/thoas/stats
	s := stats.New()
	e.Use(s.Handler)
	// Route
	e.Get("/stats", func(c *echo.Context) {
		c.JSON(200, s.Data())
	})

	// Serve index file
	e.Index("public/index.html")

	// Serve static files
	e.Static("/js", "public/js")

	//************//
	//   Routes   //
	//************//
	e.Post("/users", createUser)
	e.Get("/users", getUsers)
	e.Get("/users/:id", getUser)

	// Start server
	e.Run(":8080")
}

```

### Benchmark
Based on [julienschmidt/go-http-routing-benchmark] (https://github.com/vishr/go-http-routing-benchmark), March 25, 2015
##### [GitHub API](http://developer.github.com/v3)
```
BenchmarkAce_GithubAll	    	20000	     70318 ns/op	   13792 B/op	     167 allocs/op
BenchmarkBear_GithubAll	    	10000	    262758 ns/op	   79952 B/op	     943 allocs/op
BenchmarkBeego_GithubAll	 	 3000	    504264 ns/op	  146272 B/op	    2092 allocs/op
BenchmarkEcho_GithubAll	    	30000	     45349 ns/op	       0 B/op	       0 allocs/op
BenchmarkBone_GithubAll	     	 1000	   2058931 ns/op	  648016 B/op	    8119 allocs/op
BenchmarkDenco_GithubAll		20000	     79656 ns/op	   20224 B/op	     167 allocs/op
BenchmarkGin_GithubAll	    	20000	     71009 ns/op	   13792 B/op	     167 allocs/op
BenchmarkGocraftWeb_GithubAll	 5000	    384342 ns/op	  133280 B/op	    1889 allocs/op
BenchmarkGoji_GithubAll	    	 3000	    565323 ns/op	   56113 B/op	     334 allocs/op
BenchmarkGoJsonRest_GithubAll	 5000	    461591 ns/op	  135995 B/op	    2940 allocs/op
BenchmarkGoRestful_GithubAll	  200	   8828847 ns/op	  649139 B/op	    7355 allocs/op
BenchmarkGorillaMux_GithubAll	  200	   6794821 ns/op	  153137 B/op	    1791 allocs/op
BenchmarkHttpRouter_GithubAll	30000	     51688 ns/op	   13792 B/op	     167 allocs/op
BenchmarkHttpTreeMux_GithubAll	10000	    136677 ns/op	   56112 B/op	     334 allocs/op
BenchmarkKocha_GithubAll	   	10000	    145115 ns/op	   23304 B/op	     843 allocs/op
BenchmarkMacaron_GithubAll	     2000	    684714 ns/op	  224960 B/op	    2315 allocs/op
BenchmarkMartini_GithubAll	      100	  10501805 ns/op	  237953 B/op	    2686 allocs/op
BenchmarkPat_GithubAll	     	  500	   3987941 ns/op	 1504101 B/op	   32222 allocs/op
BenchmarkRevel_GithubAll	     2000	   1127175 ns/op	  345553 B/op	    5918 allocs/op
BenchmarkRivet_GithubAll	   	10000	    239793 ns/op	   84272 B/op	    1079 allocs/op
BenchmarkTango_GithubAll	      500	   3452359 ns/op	 1338664 B/op	   27736 allocs/op
BenchmarkTigerTonic_GithubAll	 2000	    951609 ns/op	  241088 B/op	    6052 allocs/op
BenchmarkTraffic_GithubAll	      200	   7496057 ns/op	 2664761 B/op	   22390 allocs/op
BenchmarkVulcan_GithubAll	     5000	    300807 ns/op	   19894 B/op	     609 allocs/op
BenchmarkZeus_GithubAll	         2000	    801269 ns/op	  300688 B/op	    2648 allocs/op
```
