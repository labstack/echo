+++
title = "Streaming Response"
description = "Streaming response example for Echo"
[menu.main]
  name = "Streaming Response"
  parent = "cookbook"
  weight = 3
+++

- Send data as it is produced
- Streaming JSON response with chunked transfer encoding

## Server

`server.go`

{{< embed "streaming-response/server.go" >}}

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

## [Source Code]({{< source "streaming-response" >}})

## Maintainers

- [vishr](https://github.com/vishr)
