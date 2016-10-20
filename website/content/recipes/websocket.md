+++
title = "WebSocket Recipe"
description = "WebSocket recipe / example for Echo"
[menu.side]
  name = "WebSocket"
  parent = "recipes"
  weight = 5
+++

## WebSocket Recipe

> Only supported in `standard` engine.

### Using `net` WebSocket

#### Server

`server.go`

{{< embed "websocket/net/server.go" >}}

### Using `gorilla` WebSocket

#### Server

`server.go`

{{< embed "websocket/gorilla/server.go" >}}

### Client

`index.html`

{{< embed "websocket/public/index.html" >}}

### Output

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

### Maintainers

- [vishr](https://github.com/vishr)

### [Source Code]({{< source "websocket" >}})
