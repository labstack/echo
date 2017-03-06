+++
title = "File Upload"
description = "File upload example for Echo"
[menu.main]
  name = "File Upload"
  parent = "cookbook"
+++

## How to upload single file with fields?

### Server

`server.go`

{{< embed "file-upload/single/server.go" >}}

### Client

`index.html`

{{< embed "file-upload/single/public/index.html" >}}

## How to upload multiple files with fields?

### Server

`server.go`

{{< embed "file-upload/multiple/server.go" >}}

### Client

`index.html`

{{< embed "file-upload/multiple/public/index.html" >}}

## Source Code

- [single]({{< source "file-upload/single" >}})
- [multiple]({{< source "file-upload/multiple" >}})

## Maintainers

- [vishr](https://github.com/vishr)
