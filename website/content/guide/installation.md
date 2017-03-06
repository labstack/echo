+++
title = "Installation"
description = "Installing Echo"
[menu.main]
  name = "Installation"
  parent = "guide"
+++

## Prerequisites

- [Install](https://golang.org/doc/install) Go
- [Set](https://golang.org/doc/code.html#GOPATH) GOPATH

## Using [go get](https://golang.org/cmd/go/#hdr-Download_and_install_packages_and_dependencies)

```sh
$ cd <project in $GOPATH>
$ go get -u github.com/labstack/echo/...
```

## Using [glide](http://glide.sh)

```sh
$ cd <project in $GOPATH>
$ glide get github.com/labstack/echo#~3.0
```

## Using [govendor](https://github.com/kardianos/govendor)

```sh
$ cd <project in $GOPATH>
$ govendor fetch github.com/labstack/echo@v3.0
```

Echo is developed using Go `1.7.x` and tested with Go `1.6.x` and `1.7.x`.
Echo follows [semantic versioning](http://semver.org) managed through GitHub
releases, specific version of Echo can be installed using a [package manager](https://github.com/avelino/awesome-go#package-management).
