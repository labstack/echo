---
title: WebSocket
menu:
  side:
    parent: recipes
    weight: 5
---

### Server

`server.go`

{{< embed "websocket/main.go" >}}

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
