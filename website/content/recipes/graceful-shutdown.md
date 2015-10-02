---
title: Graceful Shutdown
menu:
  main:
    parent: recipes
---

### With [graceful](https://github.com/tylerb/graceful)

`server.go`

```go
package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/tylerb/graceful"
)

func main() {
	// Setup
	e := echo.New()
	e.Get("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Sue sews rose on slow jor crows nose")
	})

	graceful.ListenAndServe(e.Server(":1323"), 5*time.Second)
}
```

### With [grace](https://github.com/facebookgo/grace)

`server.go`

```go
package main

import (
	"net/http"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo"
)

func main() {
	// Setup
	e := echo.New()
	e.Get("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Six sick bricks tick")
	})

	gracehttp.Serve(e.Server(":1323"))
}
```

## Source Code

[`graceful`](https://github.com/labstack/echo/blob/master/recipes/graceful-shutdown/graceful)

[`grace`](https://github.com/labstack/echo/blob/master/recipes/graceful-shutdown/grace)
