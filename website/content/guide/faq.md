+++
title = "FAQ"
description = "Frequently asked questions in Echo"
[menu.main]
  name = "FAQ"
  parent = "guide"
  weight = 20
+++

Q: How to retrieve `*http.Request` and `http.ResponseWriter` from `echo.Context`?

- `http.Request` > `c.Request()`
- `http.ResponseWriter` > `c.Response()`

Q: How to use standard handler `func(http.ResponseWriter, *http.Request)` with Echo?

```go
func handler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Echo!")
}

func main() {
	e := echo.New()
	e.GET("/", echo.WrapHandler(http.HandlerFunc(handler)))
	e.Start(":1323")
}
```

Q: How to use standard middleware `func(http.Handler) http.Handler` with Echo?

```go
func middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("middleware")
		h.ServeHTTP(w, r)
	})
}

func main() {
	e := echo.New()
	e.Use(echo.WrapMiddleware(middleware))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Echo!")
	})
	e.Start(":1323")
}
```

Q: How to run Echo on a specific IP address?

```go
e.Start("<ip>:<port>")
```

Q: How to run Echo on a random port?

```go
e.Start(":0")
```

Q: How to run Echo on a unix domain socket?

```go
e := echo.New()
e.GET("/", func(c echo.Context) error {
  return c.String(http.StatusOK, "OK")
})
os.Remove("/tmp/echo.sock")
l, err := net.Listen("unix", "/tmp/echo.sock")
if err != nil {
  e.Logger.Fatal(err)
}
e.Listener = l
e.Logger.Fatal(e.Start(""))
```

```sh
curl --unix-socket /tmp/echo.sock http://localhost
```

