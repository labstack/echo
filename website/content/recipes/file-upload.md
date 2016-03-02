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

If you just want to upload a single file:

```go
file, fh, err := req.FormFile("file")
if err != nil {
    return err
}
defer file.Close()

// Destination
dst, err := os.Create(fh.Filename)
if err != nil {
    return err
}
defer dst.Close()

// Copy
if _, err = io.Copy(dst, file); err != nil {
    return err
}
```

### Server

`server.go`

{{< embed "file-upload/server.go" >}}

### Client

`index.html`

{{< embed "file-upload/public/index.html" >}}

### Maintainers

- [vishr](https://github.com/vishr)

### [Source Code](https://github.com/vishr/recipes/blob/master/echo/recipes/file-upload)
