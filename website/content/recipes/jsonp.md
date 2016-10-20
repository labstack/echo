+++
title = "JSONP Recipe"
description = "JSONP recipe / example for Echo"
[menu.side]
  name = "JSONP"
  parent = "recipes"
  weight = 6
+++

## JSONP Recipe

JSONP is a method that allows cross-domain server calls. You can read more about it at the JSON versus JSONP Tutorial.

### Server

`server.go`

{{< embed "jsonp/server.go" >}}

### Client

`index.html`

{{< embed "jsonp/public/index.html" >}}

### Maintainers

- [willf](https://github.com/willf)

### [Source Code]({{< source "jsonp" >}})
