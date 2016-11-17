+++
title = "WebSocket Example"
description = "WebSocket example for Echo"
[menu.side]
  name = "WebSocket"
  parent = "recipes"
  weight = 5
+++

## Using `net` WebSocket

### Server

`server.go`

{{< embed "websocket/net/server.go" >}}

## Using `gorilla` WebSocket

### Server

`server.go`

{{< embed "websocket/gorilla/server.go" >}}

## Client

`index.html`

{{< embed "websocket/public/index.html" >}}

## Output

`Client`

```sh
Hello, Client!
Hello, Client!
Hello, Client!
Hello, Client!
Hello, Client!
```

`Server`

```sh
Hello, Server!
Hello, Server!
Hello, Server!
Hello, Server!
Hello, Server!
```

## [Source Code]({{< source "websocket" >}})

## Maintainers

- [vishr](https://github.com/vishr)
