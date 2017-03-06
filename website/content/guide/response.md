+++
title = "Response"
description = "Sending HTTP response in Echo"
[menu.main]
  name = "Response"
  parent = "guide"
+++

## Send String

`Context#String(code int, s string)` can be used to send plain text response with status
code.

*Example*

```go
func(c echo.Context) error {
  return c.String(http.StatusOK, "Hello, World!")
}
```

## Send HTML (Reference to templates)

`Context#HTML(code int, html string)` can be used to send simple HTML response with
status code. If you are looking to send dynamically generate HTML see [templates](/guide/templates).

*Example*

```go
func(c echo.Context) error {
  return c.HTML(http.StatusOK, "<strong>Hello, World!</strong>")
}
```

### Send HTML Blob

`Context#HTMLBlob(code int, b []byte)` can be used to send HTML blob with status
code. You may find it handy using with a template engine which outputs `[]byte`.

## Render Template

[Learn more](/guide/templates)

## Send JSON

`Context#JSON(code int, i interface{})` can be used to encode a provided Go type into
JSON and send it as response with status code.

*Example*

```go
// User
type User struct {
  Name  string `json:"name" xml:"name"`
  Email string `json:"email" xml:"email"`
}

// Handler
func(c echo.Context) error {
  u := &User{
    Name:  "Jon",
    Email: "jon@labstack.com",
  }
  return c.JSON(http.StatusOK, u)
}
```

### Stream JSON

`Context#JSON()` internally uses `json.Marshal` which may not be efficient to large JSON,
in that case you can directly stream JSON.

*Example*

```go
func(c echo.Context) error {
  u := &User{
    Name:  "Jon",
    Email: "jon@labstack.com",
  }
  c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
  c.Response().WriteHeader(http.StatusOK)
  return json.NewEncoder(c.Response()).Encode(u)
}
```

### JSON Pretty

`Context#JSONPretty(code int, i interface{}, indent string)` can be used to a send
a JSON response which is pretty printed based on indent, which could spaces or tabs.

Example below sends a pretty print JSON indented with spaces:

```go
func(c echo.Context) error {
  u := &User{
    Name:  "Jon",
    Email: "joe@labstack.com",
  }
  return c.JSONPretty(http.StatusOK, u, "  ")
}
```

```js
{
  "email": "joe@labstack.com",
  "name": "Jon"
}
```

### JSON Blob

`Context#JSONBlob(code int, b []byte)` can be used to send pre-encoded JSON blob directly
from external source, for example, database.

*Example*

```go
func(c echo.Context) error {
  encodedJSON := []byte{} // Encoded JSON from external source
  return c.JSONBlob(http.StatusOK, encodedJSON)
}
```

## Send JSONP

`Context#JSONP(code int, callback string, i interface{})` can be used to encode a provided
Go type into JSON and send it as JSONP payload constructed using a callback, with
status code.

[*Example*](/cookbook/jsonp)

## Send XML

`Context#XML(code int, i interface{})` can be used to encode a provided Go type into
XML and send it as response with status cod.

*Example*

```go
func(c echo.Context) error {
  u := &User{
    Name:  "Jon",
    Email: "jon@labstack.com",
  }
  return c.XML(http.StatusOK, u)
}
```

### Stream XML

`Context#XML` internally uses `xml.Marshal` which may not be efficient to large XML,
in that case you can directly stream XML.

*Example*

```go
func(c echo.Context) error {
  u := &User{
    Name:  "Jon",
    Email: "jon@labstack.com",
  }
  c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationXMLCharsetUTF8)
  c.Response().WriteHeader(http.StatusOK)
  return xml.NewEncoder(c.Response()).Encode(u)
}
```

### XML Pretty

`Context#XMLPretty(code int, i interface{}, indent string)` can be used to a send
an XML response which is pretty printed based on indent, which could spaces or tabs.

Example below sends a pretty print XML indented with spaces:

```go
func(c echo.Context) error {
  u := &User{
    Name:  "Jon",
    Email: "joe@labstack.com",
  }
  return c.XMLPretty(http.StatusOK, u, "  ")
}
```

```xml
<?xml version="1.0" encoding="UTF-8"?>
<User>
  <Name>Jon</Name>
  <Email>joe@labstack.com</Email>
</User>
```

### XML Blob

`Context#XMLBlob(code int, b []byte)` can be used to send pre-encoded XML blob directly
from external source, for example, database.

*Example*

```go
func(c echo.Context) error {
  encodedXML := []byte{} // Encoded XML from external source
  return c.XMLBlob(http.StatusOK, encodedXML)
}
```

## Send File

`Context#File(file string)` can be used to send the content of file as response.
It automatically sets the correct content type and handles caching gracefully.

*Example*

```go
func(c echo.Context) error {
  return c.File("<PATH_TO_YOUR_FILE>")
}
```

## Send Attachment

`Context#Attachment(file, name string)` is similar to `File()` except that it is
used to send file as attachment with provided name.

*Example*

```go
func(c echo.Context) error {
  return c.Attachment("<PATH_TO_YOUR_FILE>")
}
```

## Send Inline

`Context#Inline(file, name string)` is similar to `File()` except that it is
used to send file as inline with provided name.

*Example*

```go
func(c echo.Context) error {
  return c.Inline("<PATH_TO_YOUR_FILE>")
}
```

## Send Blob

`Context#Blob(code int, contentType string, b []byte)` can be used to send an arbitrary
data response with provided content type and status code.

*Example*

```go
func(c echo.Context) (err error) {
  data := []byte(`0306703,0035866,NO_ACTION,06/19/2006
	  0086003,"0005866",UPDATED,06/19/2006`)
	return c.Blob(http.StatusOK, "text/csv", data)
}
```

// Stream sends a streaming response with status code and content type.
		Stream(code int, contentType string, r io.Reader) error

## Send Stream

`Context#Stream(code int, contentType string, r io.Reader)` can be used to send an
arbitrary data stream response with provided content type, `io.Reader` and status
code.

*Example*

```go
func(c echo.Context) error {
  f, err := os.Open("<PATH_TO_IMAGE>")
  if err != nil {
    return err
  }
  return c.Stream(http.StatusOK, "image/png", f)
}
```

## Send No Content

`Context#NoContent(code int)` can be used to send empty body with status code.

*Example*

```go
func(c echo.Context) error {
  return c.NoContent(http.StatusOK)
}
```

## Redirect Request

`Context#Redirect(code int, url string)` can be used to redirect the request to
a provided URL with status code.

*Example*

```go
func(c echo.Context) error {
  return c.Redirect(http.StatusMovedPermanently, "<URL>")
}
```
