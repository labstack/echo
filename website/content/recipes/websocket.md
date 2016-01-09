---
title: WebSocket
menu:
  side:
    parent: recipes
    weight: 5
---

### Server

`server.go`

{{< embed "websocket/server.go" >}}

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

### [Source Code](https://github.com/vishr/recipes/blob/master/echo/recipes/websocket)
