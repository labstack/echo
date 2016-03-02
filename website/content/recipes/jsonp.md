---
title: JSONP
menu:
  side:
    parent: recipes
    weight: 6
---

JSONP is a method that allows cross-domain server calls. You can read more about it at the JSON versus JSONP Tutorial.

### Server

`server.go`

{{< embed "jsonp/server.go" >}}

### Client

`index.html`

{{< embed "jsonp/public/index.html" >}}

### Maintainers

- [willf](https://github.com/willf)

### [Source Code](https://github.com/vishr/recipes/blob/master/echo/recipes/jsonp)
