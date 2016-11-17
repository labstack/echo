+++
title = "Graceful Shutdown Example"
description = "Graceful shutdown example for Echo"
[menu.side]
  name = "Graceful Shutdown"
  parent = "recipes"
  weight = 13
+++

Echo now ships with graceful server termination inside it, to accomplish it Echo
uses `github.com/tylerb/graceful` library. By Default echo uses 15 seconds as shutdown
timeout, giving 15 secs to open connections at the time the server starts to shut-down.
In order to change this default 15 seconds you could change the `ShutdownTimeout`
property of your Echo instance as needed by doing something like:

`server.go`

{{< embed "graceful-shutdown/server.go" >}}

## Source Code

- [graceful]({{< source "graceful-shutdown/graceful" >}})

## Maintainers

- [mertenvg](https://github.com/mertenvg)
- [apaganobeleno](https://github.com/apaganobeleno)
