+++
title = "HTTP/2 Example"
description = "HTTP/2 example for Echo"
[menu.side]
  name = "HTTP/2"
  parent = "recipes"
  weight = 3
+++

## What is HTTP/2?

HTTP/2 (originally named HTTP/2.0) is the second major version of the HTTP network
protocol used by the World Wide Web

### Features

- Binary, instead of textual.
- Fully multiplexed, instead of ordered and blocking, can therefore use just one TCP connection.
- Uses header compression to reduce overhead.
- Allows servers to "push" responses proactively into client caches.

## How to run an HTTP/2 and HTTPS server?

### Generate a self-signed X.509 TLS certificate (HTTP/2 requires TLS to operate)

```sh
go run $GOROOT/src/crypto/tls/generate_cert.go --host localhost
```

This will generate `cert.pem` and `key.pem` files.

> For demo purpose, we are using a self-signed certificate. Ideally you should obtain
a certificate from [CA](https://en.wikipedia.org/wiki/Certificate_authority).

### Configure a server with `engine.Config`

`server.go`

{{< embed "http2/server.go" >}}

### Endpoints

- https://localhost:1323/request (Displays the information about received HTTP request)
- https://localhost:1323/stream (Streams the current time every second)

## [Source Code]({{< source "http2" >}})

## Maintainers

- [vishr](https://github.com/vishr)
