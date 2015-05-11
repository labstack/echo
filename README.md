# [Echo](http://labstack.github.io/echo) [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/labstack/echo) [![Build Status](http://img.shields.io/travis/labstack/echo.svg?style=flat-square)](https://travis-ci.org/labstack/echo) [![Coverage Status](http://img.shields.io/coveralls/labstack/echo.svg?style=flat-square)](https://coveralls.io/r/labstack/echo)
Echo is a fast HTTP router (zero memory allocation) and micro web framework in Go.

## Features

- Fast :rocket: HTTP router which smartly prioritize routes.
- Extensible middleware/handler, supports:
	- Middleware
		- `echo.MiddlewareFunc`
		- `func(echo.HandlerFunc) echo.HandlerFunc`
		- `func(*echo.Context) *echo.HTTPError`
		- `func(*echo.Context)`
		- `func(http.Handler) http.Handler`
		- `http.Handler`
		- `http.HandlerFunc`
		- `func(http.ResponseWriter, *http.Request)`
		- `func(http.ResponseWriter, *http.Request) *echo.HTTPError`
	- Handler
		- `echo.HandlerFunc`
		- `func(*echo.Context) *echo.HTTPError`
		- `func(*echo.Context)`
		- `http.Handler`
		- `http.HandlerFunc`
		- `func(http.ResponseWriter, *http.Request)`
		- `func(http.ResponseWriter, *http.Request) *echo.HTTPError`
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

##[Hello, World!](https://github.com/labstack/echo/tree/master/examples/hello)

```go
package main

import (
	"net/http"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

// Handler
func hello(c *echo.Context) *echo.HTTPError {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(mw.Logger)

	// Routes
	e.Get("/", hello)

	// Start server
	e.Run(":1323)
}
```

##[Examples](https://github.com/labstack/echo/tree/master/examples)

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
