+++
title = "Graceful Shutdown Recipe"
description = "Graceful shutdown recipe / example for Echo"
[menu.side]
  name = "Graceful Shutdown"
  parent = "recipes"
  weight = 13
+++

## Graceful Shutdown Recipe

### Using [grace](https://github.com/facebookgo/grace)

`server.go`

{{< embed "graceful-shutdown/grace/server.go" >}}

### Using [graceful](https://github.com/tylerb/graceful)

`server.go`

{{< embed "graceful-shutdown/graceful/server.go" >}}

### Maintainers

- [mertenvg](https://github.com/mertenvg)

### Source Code

- [graceful]({{< source "graceful-shutdown/graceful" >}})
- [grace]({{< source "graceful-shutdown/grace" >}})
