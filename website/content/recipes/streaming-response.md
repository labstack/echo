---
title: Streaming Response
menu:
  main:
    parent: recipes
---

- Send data as it is produced
- Streaming JSON response with chunked transfer encoding

## Server

`server.go`

```go
package main

import (
	"net/http"
	"time"

	"encoding/json"

	"github.com/labstack/echo"
)

type (
	Geolocation struct {
		Altitude  float64
		Latitude  float64
		Longitude float64
	}
)

var (
	locations = []Geolocation{
		{-97, 37.819929, -122.478255},
		{1899, 39.096849, -120.032351},
		{2619, 37.865101, -119.538329},
		{42, 33.812092, -117.918974},
		{15, 37.77493, -122.419416},
	}
)

func main() {
	e := echo.New()
	e.Get("/", func(c *echo.Context) error {
		c.Response().Header().Set(echo.ContentType, echo.ApplicationJSON)
		c.Response().WriteHeader(http.StatusOK)
		for _, l := range locations {
			if err := json.NewEncoder(c.Response()).Encode(l); err != nil {
				return err
			}
			c.Response().Flush()
			time.Sleep(1 * time.Second)
		}
		return nil
	})
	e.Run(":1323")
}
```

## Client

```sh
$ curl localhost:1323
```

## Output

```sh
{"Altitude":-97,"Latitude":37.819929,"Longitude":-122.478255}
{"Altitude":1899,"Latitude":39.096849,"Longitude":-120.032351}
{"Altitude":2619,"Latitude":37.865101,"Longitude":-119.538329}
{"Altitude":42,"Latitude":33.812092,"Longitude":-117.918974}
{"Altitude":15,"Latitude":37.77493,"Longitude":-122.419416}
```

## [Source Code](https://github.com/labstack/echo/blob/master/recipes/streaming-response)
