+++
title = "File Upload Recipe"
description = "File upload recipe / example for Echo"
[menu.side]
  name = "File Upload"
  parent = "recipes"
  weight = 7
+++

## File Upload Recipe

### How to upload single file with fields?

#### Server

`server.go`

{{< embed "file-upload/single/server.go" >}}

#### Client

`index.html`

{{< embed "file-upload/single/public/index.html" >}}

### How to upload multiple files with fields?

#### Server

`server.go`

{{< embed "file-upload/multiple/server.go" >}}

#### Client

`index.html`

{{< embed "file-upload/multiple/public/index.html" >}}

### Maintainers

- [vishr](https://github.com/vishr)

### Source Code

- [single]({{< source "file-upload/single" >}})
- [multiple]({{< source "file-upload/multiple" >}})
