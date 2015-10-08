---
title: File Upload
menu:
  side:
    parent: recipes
    weight: 7
---

- Multipart/form-data file upload
- Multiple form fields and files

Use `req.ParseMultipartForm(16 << 20)` for manually parsing multipart form. It gives
us an option to specify the maximum memory used while parsing the request body.

### Server

`server.go`

{{< embed "file-upload/server.go" >}}

### Client

`index.html`

{{< embed "file-upload/public/index.html" >}}

### Maintainers

- [vishr](https://github.com/vishr)

### [Source Code](https://github.com/labstack/echo/blob/master/recipes/file-upload)
