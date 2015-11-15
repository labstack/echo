# [Echo](http://labstack.com/echo) [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/labstack/echo) [![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/labstack/echo/master/LICENSE) [![Build Status](http://img.shields.io/travis/labstack/echo.svg?style=flat-square)](https://travis-ci.org/labstack/echo) [![Coverage Status](http://img.shields.io/coveralls/labstack/echo.svg?style=flat-square)](https://coveralls.io/r/labstack/echo) [![Join the chat at https://gitter.im/labstack/echo](https://img.shields.io/badge/gitter-join%20chat-brightgreen.svg?style=flat-square)](https://gitter.im/labstack/echo)

A fast and unfancy micro web framework for Go.

## Features

- Fast HTTP router which smartly prioritize routes.
- Extensible middleware, supports:
	- `echo.MiddlewareFunc`
	- `func(echo.HandlerFunc) echo.HandlerFunc`
	- `echo.HandlerFunc`
	- `func(*echo.Context) error`
	- `func(http.Handler) http.Handler`
	- `http.Handler`
	- `http.HandlerFunc`
	- `func(http.ResponseWriter, *http.Request)`
- Extensible handler, supports:
    - `echo.HandlerFunc`
    - `func(*echo.Context) error`
    - `http.Handler`
    - `http.HandlerFunc`
    - `func(http.ResponseWriter, *http.Request)`
- Sub-router/Groups
- Handy functions to send variety of HTTP response:
    - HTML
    - HTML via templates
    - String
    - JSON
    - JSONP
    - XML
    - File
    - NoContent
    - Redirect
    - Error
- Build-in support for:
	- Favicon
	- Index file
	- Static files
	- WebSocket
- Centralized HTTP error handling.
- Customizable HTTP request binding function.
- Customizable HTTP response rendering function, allowing you to use any HTML template engine.

## Performance

Based on [vishr/go-http-routing-benchmark] (https://github.com/vishr/go-http-routing-benchmark), June 5, 2015.

##### [GitHub API](http://developer.github.com/v3)

> Echo: 38662 ns/op, 0 B/op, 0 allocs/op

![Performance](http://i.imgur.com/hB2qdRS.png)

```
BenchmarkAce_GithubAll              20000             93675 ns/op           13792 B/op      167 allocs/op
BenchmarkBear_GithubAll             10000            264194 ns/op           79952 B/op      943 allocs/op
BenchmarkBeego_GithubAll             2000           1109160 ns/op          146272 B/op      2092 allocs/op
BenchmarkBone_GithubAll              1000           2063973 ns/op          648016 B/op     8119 allocs/op
BenchmarkDenco_GithubAll            20000             83114 ns/op           20224 B/op       167 allocs/op
BenchmarkEcho_GithubAll             30000             38662 ns/op               0 B/op        0 allocs/op
BenchmarkGin_GithubAll              30000             43467 ns/op               0 B/op        0 allocs/op
BenchmarkGocraftWeb_GithubAll        5000            386829 ns/op          133280 B/op      1889 allocs/op
BenchmarkGoji_GithubAll              3000            561131 ns/op           56113 B/op      334 allocs/op
BenchmarkGoJsonRest_GithubAll        3000            490789 ns/op          135995 B/op      2940 allocs/op
BenchmarkGoRestful_GithubAll          100          15569513 ns/op          797239 B/op      7725 allocs/op
BenchmarkGorillaMux_GithubAll         200           7431130 ns/op          153137 B/op      1791 allocs/op
BenchmarkHttpRouter_GithubAll       30000             51192 ns/op           13792 B/op       167 allocs/op
BenchmarkHttpTreeMux_GithubAll      10000            138164 ns/op           56112 B/op       334 allocs/op
BenchmarkKocha_GithubAll            10000            139625 ns/op           23304 B/op       843 allocs/op
BenchmarkMacaron_GithubAll           2000            709932 ns/op          224960 B/op      2315 allocs/op
BenchmarkMartini_GithubAll            100          10261331 ns/op          237953 B/op      2686 allocs/op
BenchmarkPat_GithubAll                500           3989686 ns/op         1504104 B/op    32222 allocs/op
BenchmarkPossum_GithubAll            5000            259165 ns/op           97441 B/op       812 allocs/op
BenchmarkR2router_GithubAll         10000            240345 ns/op           77328 B/op      1182 allocs/op
BenchmarkRevel_GithubAll             2000           1203336 ns/op          345554 B/op      5918 allocs/op
BenchmarkRivet_GithubAll            10000            247213 ns/op           84272 B/op      1079 allocs/op
BenchmarkTango_GithubAll             5000            379960 ns/op           87081 B/op      2470 allocs/op
BenchmarkTigerTonic_GithubAll        2000            931401 ns/op          241089 B/op      6052 allocs/op
BenchmarkTraffic_GithubAll            200           7292170 ns/op         2664770 B/op     22390 allocs/op
BenchmarkVulcan_GithubAll            5000            271682 ns/op           19894 B/op       609 allocs/op
BenchmarkZeus_GithubAll              2000            748827 ns/op          300688 B/op     2648 allocs/op
```

## Installation

```sh
$ go get github.com/labstack/echo
```

## [Recipes](http://labstack.com/echo/recipes/hello-world)

## [Guide](http://labstack.com/echo/guide/installation)

## Echo System

Community created packages for Echo

- [echo-logrus](https://github.com/deoxxa/echo-logrus)
- [go_middleware](https://github.com/rightscale/go_middleware)
- [permissions2](https://github.com/xyproto/permissions2)
- [permissionbolt](https://github.com/xyproto/permissionbolt)
- [echo-middleware](https://github.com/syntaqx/echo-middleware)
- [dpecho](https://github.com/deferpanic/dpecho)
- [echosentry](https://github.com/01walid/echosentry)

## Contribute

**Use issues for everything**

- Report problems
- Discuss before sending a pull request
- Suggest new features/recipes
- Improve/fix documentation

## Credits
- [Vishal Rana](https://github.com/vishr) - Author
- [Nitin Rana](https://github.com/nr17) - Consultant
- [Contributors](https://github.com/labstack/echo/graphs/contributors)
