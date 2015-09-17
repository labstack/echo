---
title: go.rice integration
menu:
  side:
    identifier: "recipe-rice"
    parent: recipes
    weight: 14
---

[go.rice](https://github.com/GeertJohan/go.rice) is a library that can be used
to package the assets (js, css, etc) inside the binary file, so your app
can still be a single binary.

This folder contains a simple example serving an `index.html` file and a simple
`.js` file with go.rice.

### Server

`server.go`

{{< embed "rice/server.go" >}}

### Maintainers

- [vishr](https://github.com/caarlos0)

### [Source Code](https://github.com/labstack/echo/blob/master/recipes/rice)
