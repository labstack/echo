---
title: JSONP
menu:
  main:
    parent: recipes
---

JSONP is a method that allows cross-domain server calls. You can read more about it at the JSON versus JSONP Tutorial.

### Server

`server.go`

```go
package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

func main() {
	// Setup
	e := echo.New()
	e.ServeDir("/", "public")

	e.Get("/jsonp", func(c *echo.Context) error {
		callback := c.Query("callback")
		var content struct {
			Response  string    `json:"response"`
			Timestamp time.Time `json:"timestamp"`
			Random    int       `json:"random"`
		}
		content.Response = "Sent via JSONP"
		content.Timestamp = time.Now().UTC()
		content.Random = rand.Intn(1000)
		return c.JSONP(http.StatusOK, callback, &content)
	})

	// Start server
	e.Run(":3999")
}
```

### Client

`index.html`

```html
<!DOCTYPE html>
<html>

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">
    <title>JSONP</title>
    <script type="text/javascript" src="//ajax.googleapis.com/ajax/libs/jquery/1/jquery.min.js"></script>
    <script type="text/javascript">
        var host_prefix = 'http://localhost:3999';
        $(document).ready(function() {
            // JSONP version - add 'callback=?' to the URL - fetch the JSONP response to the request
            $("#jsonp-button").click(function(e) {
                e.preventDefault();
                // The only difference on the client end is the addition of 'callback=?' to the URL
                var url = host_prefix + '/jsonp?callback=?';
                $.getJSON(url, function(jsonp) {
                    console.log(jsonp);
                    $("#jsonp-response").html(JSON.stringify(jsonp, null, 2));
                });
            });
        });
    </script>

</head>

<body>
    <div class="container" style="margin-top: 50px;">
        <input type="button" class="btn btn-primary btn-lg" id="jsonp-button" value="Get JSONP response">
        <p>
            <pre id="jsonp-response"></pre>
        </p>
    </div>
</body>

</html>
```

### Maintainers

- [willf](http://github.com/willf)

### [Source Code](https://github.com/labstack/echo/blob/master/recipes/jsonp)
