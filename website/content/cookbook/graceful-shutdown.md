+++
title = "Graceful Shutdown"
description = "Graceful shutdown example for Echo"
[menu.main]
  name = "Graceful Shutdown"
  parent = "cookbook"
  weight = 13
+++

## Using [grace](https://github.com/facebookgo/grace)

`server.go`

{{< embed "graceful-shutdown/grace/server.go" >}}

## Using [graceful](https://github.com/tylerb/graceful)

`server.go`

{{< embed "graceful-shutdown/graceful/server.go" >}}

## Source Code

- [graceful]({{< source "graceful-shutdown/graceful" >}})
- [grace]({{< source "graceful-shutdown/grace" >}})

## Maintainers

- [mertenvg](https://github.com/mertenvg)
- [apaganobeleno](https://github.com/apaganobeleno)
- [vishr](https://github.com/vishr)
